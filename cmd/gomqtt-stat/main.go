package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

var urlString = flag.String("url", "tcp://0.0.0.0:1883", "broker url")

func main() {
	flag.Parse()

	mqttURL, err := url.Parse(*urlString)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Start analisys of %s\n", *urlString)

	go func() {
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

		<-finish
		fmt.Println("Closing...")
		os.Exit(0)
	}()

	var received int32

	c := client.New()

	c.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&received, 1)

		return nil
	}

	cf, err := c.Connect(client.NewConfig(*urlString))
	if err != nil {
		panic(err)
	}

	err = cf.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	sf, err := c.Subscribe(mqttURL.Path, 0)
	if err != nil {
		panic(err)
	}

	err = sf.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	var iterations int32
	var total int32

	for {
		time.Sleep(1 * time.Second)

		curReceived := atomic.LoadInt32(&received)
		atomic.StoreInt32(&received, 0)
		total += curReceived

		iterations++

		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Average Throughput: %d msg/s)\n", total/iterations)
	}
}
