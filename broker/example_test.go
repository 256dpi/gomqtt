package broker

import (
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

func Example() {
	server, err := transport.Launch("tcp://localhost:8080")
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})

	backend := NewMemoryBackend()
	backend.Logger = func(e LogEvent, c *Client, pkt packet.Generic, msg *packet.Message, err error) {
		if err != nil {
			fmt.Printf("B [%s] %s\n", e, err.Error())
		} else if msg != nil {
			fmt.Printf("B [%s] %s\n", e, msg.String())
		} else if pkt != nil {
			fmt.Printf("B [%s] %s\n", e, pkt.String())
		} else {
			fmt.Printf("B [%s]\n", e)
		}

		if e == LostConnection {
			close(done)
		}
	}

	engine := NewEngine(backend)
	engine.Accept(server)

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		fmt.Printf("C [message] %s\n", msg.String())
		close(wait)
		return nil
	}

	cf, err := c.Connect(client.NewConfig("tcp://localhost:8080"))
	if err != nil {
		panic(err)
	}

	err = cf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	sf, err := c.Subscribe("test", 0)
	if err != nil {
		panic(err)
	}

	err = sf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	pf, err := c.Publish("test", []byte("test"), 0, false)
	if err != nil {
		panic(err)
	}

	err = pf.Wait(10 * time.Second)
	if err != nil {
		panic(err)
	}

	<-wait

	err = c.Disconnect()
	if err != nil {
		panic(err)
	}

	<-done

	err = server.Close()
	if err != nil {
		panic(err)
	}

	engine.Close()

	// Output:
	// B [new connection]
	// B [packet received] <Connect ClientID="" KeepAlive=30 Username="" Password="" CleanSession=true Will=nil Version=4>
	// B [packet sent] <Connack SessionPresent=false ReturnCode=0>
	// B [packet received] <Subscribe ID=1 Subscriptions=["test"=>0]>
	// B [packet sent] <Suback ID=1 ReturnCodes=[0]>
	// B [packet received] <Publish ID=0 Message=<Message Topic="test" QOS=0 Retain=false Payload=74657374> Dup=false>
	// B [message published] <Message Topic="test" QOS=0 Retain=false Payload=74657374>
	// B [message dequeued] <Message Topic="test" QOS=0 Retain=false Payload=74657374>
	// B [packet sent] <Publish ID=0 Message=<Message Topic="test" QOS=0 Retain=false Payload=74657374> Dup=false>
	// B [message forwarded] <Message Topic="test" QOS=0 Retain=false Payload=74657374>
	// C [message] <Message Topic="test" QOS=0 Retain=false Payload=74657374>
	// B [packet received] <Disconnect>
	// B [client disconnected]
	// B [lost connection]
}
