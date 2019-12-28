package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/beorn7/perks/quantile"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "broker url")
var topic = flag.String("topic", "speed-test", "the used topic")
var qos = flag.Uint("qos", 0, "the qos level")
var wait = flag.Int("wait", 0, "time to wait in milliseconds")

var received = make(chan time.Time)

func main() {
	// parse flags
	flag.Parse()

	// create client
	cl := client.New()

	// set callback
	cl.Callback = func(msg *packet.Message, err error) error {
		// check error
		if err != nil {
			panic(err)
		}

		// queue time
		received <- time.Now()

		return nil
	}

	// connect
	cf, err := cl.Connect(client.NewConfig(*broker))
	if err != nil {
		panic(err)
	}

	// await connack
	err = cf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	// subscribe
	sf, err := cl.Subscribe(*topic, packet.QOS(*qos))
	if err != nil {
		panic(err)
	}

	// await suback
	err = sf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	// prepare quantile
	q := quantile.NewTargeted(map[float64]float64{
		0.50: 0.005,
		0.90: 0.001,
		0.99: 0.0001,
	})

	for {
		// get time
		t1 := time.Now()

		// publish message
		pf, err := cl.Publish(*topic, []byte(*topic), packet.QOS(*qos), false)
		if err != nil {
			panic(err)
		}

		// await ack
		err = pf.Wait(10 * time.Second)
		if err != nil {
			panic(err)
		}

		// await time
		t2 := <-received

		// add measurement
		q.Insert(float64(t2.Sub(t1).Nanoseconds() / 1000 / 1000))

		// get statistics
		q50 := time.Duration(q.Query(0.50)) * time.Millisecond
		q90 := time.Duration(q.Query(0.90)) * time.Millisecond
		q99 := time.Duration(q.Query(0.99)) * time.Millisecond

		// print statistics
		fmt.Printf("[%d] 0.50: %s, 0.90: %s, 0.99: %s \n", q.Count(), q50, q90, q99)

		// wait some time
		if *wait > 0 {
			time.Sleep(time.Duration(*wait) * time.Millisecond)
		}
	}
}
