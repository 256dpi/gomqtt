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

	engine := NewEngine()
	engine.Accept(server)

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			panic(err)
		}

		fmt.Println(msg.String())
		close(wait)
		return nil
	}

	cf, err := c.Connect(client.NewConfig("tcp://localhost:8080"))
	if err != nil {
		panic(err)
	}

	cf.Wait(10 * time.Second)

	sf, err := c.Subscribe("test", 0)
	if err != nil {
		panic(err)
	}

	sf.Wait(10 * time.Second)

	pf, err := c.Publish("test", []byte("test"), 0, false)
	if err != nil {
		panic(err)
	}

	pf.Wait(10 * time.Second)

	<-wait

	err = c.Disconnect()
	if err != nil {
		panic(err)
	}

	err = server.Close()
	if err != nil {
		panic(err)
	}

	engine.Close()

	// Output:
	// <Message Topic="test" QOS=0 Retain=false Payload=[116 101 115 116]>
}
