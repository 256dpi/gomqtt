package main

import (
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

var broker = flag.String("broker", "tcp://0.0.0.0:1883", "the broker url")
var filter = flag.String("filter", "#", "the filter subscription")

func main() {
	// parse flags
	flag.Parse()

	// print info
	fmt.Printf("Starting analisys of %s with filter %s...\n", *broker, *filter)

	// prepare counter
	var received int32

	// create client
	c := client.New()

	// set callback
	c.Callback = func(msg *packet.Message, err error) error {
		// check error
		if err != nil {
			panic(err)
		}

		// increment
		atomic.AddInt32(&received, 1)

		return nil
	}

	// connect
	cf, err := c.Connect(client.NewConfig(*broker))
	if err != nil {
		panic(err)
	}

	// await connack
	err = cf.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// subscribe
	sf, err := c.Subscribe(*filter, 0)
	if err != nil {
		panic(err)
	}

	// await suback
	err = sf.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// prepare counters
	var iterations int32
	var total int32

	for {
		// wait a second
		time.Sleep(1 * time.Second)

		// get and reset counter
		curReceived := atomic.LoadInt32(&received)
		atomic.StoreInt32(&received, 0)

		// increment total and iterations
		total += curReceived
		iterations++

		// print statistics
		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Average Throughput: %d msg/s)\n", total/iterations)
	}
}
