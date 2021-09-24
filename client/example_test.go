package client

import (
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/packet"
)

func ExampleClient() {
	done := make(chan struct{})

	c := New()

	c.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s: %s\n", msg.Topic, msg.Payload)
		close(done)

		return nil
	}

	config := NewConfigWithClientID("mqtt://0.0.0.0", "gomqtt/client")

	connectFuture, err := c.Connect(config)
	if err != nil {
		panic(err)
	}

	err = connectFuture.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	subscribeFuture, err := c.Subscribe("test", 0)
	if err != nil {
		panic(err)
	}

	err = subscribeFuture.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	publishFuture, err := c.Publish("test", []byte("test"), 0, false)
	if err != nil {
		panic(err)
	}

	err = publishFuture.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	<-done

	err = c.Disconnect()
	if err != nil {
		panic(err)
	}

	// Output:
	// test: test
}

func ExampleService() {
	wait := make(chan struct{})
	done := make(chan struct{})

	config := NewConfigWithClientID("mqtt://0.0.0.0", "gomqtt/service")
	config.CleanSession = false

	s := NewService()

	s.OnlineCallback = func(resumed bool) {
		fmt.Println("online!")
		fmt.Printf("resumed: %v\n", resumed)
	}

	s.OfflineCallback = func() {
		fmt.Println("offline!")
		close(done)
	}

	s.MessageCallback = func(msg *packet.Message) error {
		fmt.Printf("message: %s - %s\n", msg.Topic, msg.Payload)
		close(wait)
		return nil
	}

	err := ClearSession(config, 1*time.Second)
	if err != nil {
		panic(err)
	}

	s.Start(config)

	err = s.Subscribe("test", 0).Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	err = s.Publish("test", []byte("test"), 0, false).Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	<-wait

	s.Stop(true)

	<-done

	// Output:
	// online!
	// resumed: false
	// message: test - test
	// offline!
}
