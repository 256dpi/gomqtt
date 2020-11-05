package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"gopkg.in/tomb.v2"

	"github.com/256dpi/gomqtt/client"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "the broker url")

var numTopics = flag.Int("topic", 10, "the number of topics")
var numRobots = flag.Int("robots", 7, "the number of robots")
var numSubs = flag.Int("subs", 2, "the number of subscriptions")
var timeout = flag.Duration("timeout", 5*time.Second, "the publish timeout")

func main() {
	// parse flags
	flag.Parse()

	// print info
	fmt.Println("Launching robots...")

	// generate topics
	topics := generateTopics(*numTopics)

	// run robots
	fmt.Println("launching robots...")
	var robots []*robot
	for i := 1; i <= *numRobots; i++ {
		robots = append(robots, runRobot(*broker, strconv.Itoa(i), *numSubs, topics, *timeout))
	}

	// wait until interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	// close all robots
	fmt.Println("closing...")
	for _, r := range robots {
		r.kill()
	}
}

func generateTopics(num int) []string {
	// prepare topic levels
	l1 := make([]string, 0, num)
	l2 := make([]string, 0, num)
	l3 := make([]string, 0, num)

	// generate segments
	for i := 0; i < num; i++ {
		l1 = append(l1, randomdata.Adjective())
		l2 = append(l2, randomdata.Adjective())
		l3 = append(l3, randomdata.Adjective())
	}

	// prepare full topics
	topics := make([]string, 0, num)

	// combine levels
	for i := 0; i < num; i++ {
		topics = append(topics, strings.Join([]string{
			l1[rand.Intn(num)],
			l2[rand.Intn(num)],
			l3[rand.Intn(num)],
		}, "/"))
	}

	return topics
}

func randPick(list []string) string {
	return list[rand.Intn(len(list))]
}

type robot struct {
	topics  []string
	timeout time.Duration
	service *client.Service
	tomb    *tomb.Tomb
}

func runRobot(url, id string, subs int, topics []string, timeout time.Duration) *robot {
	// prepare config
	config := client.NewConfigWithClientID(url, "robot-"+id)

	// create service
	service := client.NewService()
	service.ResubscribeAllSubscriptions = true

	// make subscriptions
	for i := 0; i < subs; i++ {
		service.Subscribe(randPick(topics), 0)
	}

	// prepare robot with new service
	rob := &robot{
		topics:  topics,
		timeout: timeout,
		service: service,
	}

	// set online callback
	rob.service.OnlineCallback = func(resumed bool) {
		fmt.Println(config.ClientID + ": online")

		// run loop in new tomb
		rob.tomb = new(tomb.Tomb)
		rob.tomb.Go(rob.loop)
	}

	// set error callback
	rob.service.ErrorCallback = func(err error) {
		if !errors.Is(err, io.EOF) {
			fmt.Println(config.ClientID + ": " + err.Error())
		}
	}

	// set offline callback
	rob.service.OfflineCallback = func() {
		fmt.Println(config.ClientID + ": offline")

		// kill loop
		rob.tomb.Kill(nil)
		_ = rob.tomb.Wait()
		rob.tomb = nil
	}

	// start service
	rob.service.Start(config)

	return rob
}

func (r *robot) kill() {
	// stop service
	r.service.Stop(true)
}

func (r *robot) loop() error {
	for {
		// await random timeout
		select {
		case <-time.After(time.Duration(rand.Int63n(int64(r.timeout)))):
			// prepare payload
			payload := strconv.AppendInt(nil, rand.Int63n(100), 10)

			// publish message
			err := r.service.Publish(randPick(r.topics), payload, 0, false).Wait(time.Minute)
			if err != nil {
				return err
			}
		case <-r.tomb.Dying():
			return tomb.ErrDying
		}
	}
}
