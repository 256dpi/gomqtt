package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "broker url")
var pairs = flag.Int("pairs", 1, "number of pairs")
var inflight = flag.Int("inflight", 0, "available tokens")
var duration = flag.Int("duration", 0, "duration in seconds")
var length = flag.Int("length", 1, "payload length")
var retained = flag.Bool("retained", false, "retain flag")

var sent int32
var received int32
var delta int32
var total int32

var wg sync.WaitGroup

var payload []byte

func main() {
	flag.Parse()

	go func() {
		panic(http.ListenAndServe("localhost:6061", nil))
	}()

	payload = make([]byte, *length)

	fmt.Printf("Start benchmark of %s using %d pairs for %d seconds...\n", *broker, *pairs, *duration)

	go func() {
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

		<-finish
		fmt.Println("Closing...")
		os.Exit(0)
	}()

	if *duration > 0 {
		time.AfterFunc(time.Duration(*duration)*time.Second, func() {
			fmt.Println("Finishing...")
			os.Exit(0)
		})
	}

	wg.Add(*pairs * 2)

	for i := 0; i < *pairs; i++ {
		id := strconv.Itoa(i)

		var tokens chan struct{}
		if *inflight > 0 {
			tokens = make(chan struct{}, *inflight)
			for i := 0; i < *inflight; i++ {
				tokens <- struct{}{}
			}
		}

		go consumer(id, tokens)
		go publisher(id, tokens)
	}

	go reporter()

	wg.Wait()
}

func connection(id string) transport.Conn {
	conn, err := transport.Dial(*broker)
	if err != nil {
		panic(err)
	}

	conn.SetMaxWriteDelay(10 * time.Millisecond)

	mqttURL, err := url.Parse(*broker)
	if err != nil {
		panic(err)
	}

	connect := packet.NewConnect()
	connect.ClientID = "gomqtt-benchmark/" + id

	if mqttURL.User != nil {
		connect.Username = mqttURL.User.Username()
		pw, _ := mqttURL.User.Password()
		connect.Password = pw
	}

	err = conn.Send(connect, false)
	if err != nil {
		panic(err)
	}

	pkt, err := conn.Receive()
	if err != nil {
		panic(err)
	}

	if connack, ok := pkt.(*packet.Connack); ok {
		if connack.ReturnCode == packet.ConnectionAccepted {
			fmt.Printf("Connected: %s\n", id)

			return conn
		}
	}

	panic("connection failed")
}

func consumer(id string, tokens chan<- struct{}) {
	name := "consumer/" + id
	conn := connection(name)

	subscribe := packet.NewSubscribe()
	subscribe.ID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: id, QOS: 0},
	}

	err := conn.Send(subscribe, false)
	if err != nil {
		panic(err)
	}

	_, err = conn.Receive()
	if err != nil {
		panic(err)
	}

	for {
		_, err := conn.Receive()
		if err != nil {
			panic(err)
		}

		if tokens != nil {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}

		atomic.AddInt32(&received, 1)
		atomic.AddInt32(&delta, -1)
		atomic.AddInt32(&total, 1)
	}
}

func publisher(id string, tokens <-chan struct{}) {
	name := "publisher/" + id
	conn := connection(name)

	publish := packet.NewPublish()
	publish.Message.Topic = id
	publish.Message.Payload = payload
	publish.Message.Retain = *retained

	for {
		if tokens != nil {
			<-tokens
		}

		err := conn.Send(publish, true)
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&sent, 1)
		atomic.AddInt32(&delta, 1)
	}
}

func reporter() {
	var iterations int32

	for {
		time.Sleep(1 * time.Second)

		curSent := atomic.LoadInt32(&sent)
		curReceived := atomic.LoadInt32(&received)
		curDelta := atomic.LoadInt32(&delta)
		curTotal := atomic.LoadInt32(&total)

		iterations++

		fmt.Printf("Sent: %d msgs - ", curSent)
		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Buffered: %d msgs) ", curDelta)
		fmt.Printf("(Average Throughput: %d msg/s)\n", curTotal/iterations)

		atomic.StoreInt32(&sent, 0)
		atomic.StoreInt32(&received, 0)
	}
}
