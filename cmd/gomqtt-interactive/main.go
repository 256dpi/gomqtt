package main

import (
	"strconv"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/abiosoft/ishell"
)

const timeout = 5 * time.Second

func main() {
	// create new shell.
	shell := ishell.New()

	// display welcome info.
	shell.Println("Interactive MQTT Client")

	// prepare client
	var c *client.Client

	// prepare callback
	cb := func(msg *packet.Message, err error) error {
		// check error
		if err != nil {
			shell.Printf("Error: %s\n", err.Error())
			c = nil
			return nil
		}

		// print message
		shell.Printf("Message: %s\n", string(msg.Payload))
		shell.Printf("Topic: %s\n", msg.Topic)
		shell.Printf("QOS: %d\n", msg.QOS)
		shell.Printf("Retain: %t\n", msg.Retain)

		return nil
	}

	// add connect command
	shell.AddCmd(&ishell.Cmd{
		Name:     "connect",
		Help:     "connect to broker",
		LongHelp: `connect URL:string CLIENT_ID:string CLEAN:bool`,
		Func: func(ctx *ishell.Context) {
			// check state
			if c != nil {
				shell.Println("Failed: already connected")
				return
			}

			// check args
			if len(ctx.Args) == 0 {
				shell.Println("Failed: missing arguments")
				return
			}

			// get url
			url := ctx.Args[0]

			// get client id
			clientID := ""
			if len(ctx.Args) >= 2 {
				clientID = ctx.Args[1]
			}

			// get clean flag
			clean := true
			if len(ctx.Args) >= 3 {
				clean = ctx.Args[2] == "true"
			}

			// prepare config
			cfg := client.NewConfigWithClientID(url, clientID)
			cfg.CleanSession = clean

			// print config
			shell.Printf("Broker URL: %s\n", cfg.BrokerURL)
			shell.Printf("Client ID: %s\n", cfg.ClientID)
			shell.Printf("Clean Session: %t\n", cfg.CleanSession)

			// create new client
			c = client.New()
			c.Callback = cb

			// connect to broker
			cf, err := c.Connect(cfg)
			if err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}
			if err = cf.Wait(timeout); err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				shell.Printf("Return Code: %s\n", cf.ReturnCode().String())
				c = nil
				return
			}

			// print config
			shell.Println("Connected!")
			shell.Printf("Session Present: %t\n", cf.SessionPresent())

		},
	})

	// add subscribe command
	shell.AddCmd(&ishell.Cmd{
		Name:     "subscribe",
		Help:     "subscribe a topic",
		LongHelp: `subscribe TOPIC QOS`,
		Func: func(ctx *ishell.Context) {
			// check state
			if c == nil {
				shell.Println("Failed: not connected")
				return
			}

			// check args
			if len(ctx.Args) == 0 {
				shell.Println("failed: missing arguments")
				return
			}

			// get topic
			topic := ctx.Args[0]

			// get qos
			qos := 0
			if len(ctx.Args) >= 2 {
				qos, _ = strconv.Atoi(ctx.Args[1])
			}

			// subscribe
			sf, err := c.Subscribe(topic, packet.QOS(qos))
			if err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}
			if err = sf.Wait(timeout); err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				shell.Printf("Return Codes: %+v\n", sf.ReturnCodes())
				c = nil
				return
			}

			// print config
			shell.Println("Subscribed!")
		},
	})

	// add publish command
	shell.AddCmd(&ishell.Cmd{
		Name:     "publish",
		Help:     "publish a message",
		LongHelp: `publish TOPIC PAYLOAD QOS RETAIN`,
		Func: func(ctx *ishell.Context) {
			// check state
			if c == nil {
				shell.Println("Failed: not connected")
				return
			}

			// check args
			if len(ctx.Args) == 0 {
				shell.Println("failed: missing arguments")
				return
			}

			// get topic
			topic := ctx.Args[0]

			// get payload
			var payload []byte
			if len(ctx.Args) >= 2 {
				payload = []byte(ctx.Args[1])
			}

			// get qos
			qos := 0
			if len(ctx.Args) >= 3 {
				qos, _ = strconv.Atoi(ctx.Args[2])
			}

			// get retain
			retained := false
			if len(ctx.Args) >= 4 {
				retained = ctx.Args[3] == "true"
			}

			// publish
			pf, err := c.Publish(topic, payload, packet.QOS(qos), retained)
			if err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}
			if err = pf.Wait(timeout); err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}

			// print config
			shell.Println("Published!")
		},
	})

	// add unsubscribe command
	shell.AddCmd(&ishell.Cmd{
		Name:     "unsubscribe",
		Help:     "unsubscribe a topic",
		LongHelp: `unsubscribe TOPIC`,
		Func: func(ctx *ishell.Context) {
			// check state
			if c == nil {
				shell.Println("Failed: not connected")
				return
			}

			// check args
			if len(ctx.Args) == 0 {
				shell.Println("failed: missing arguments")
				return
			}

			// get topic
			topic := ctx.Args[0]

			// unsubscribe
			uf, err := c.Unsubscribe(topic)
			if err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}
			if err = uf.Wait(timeout); err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}

			// print config
			shell.Println("Unsubscribed!")
		},
	})

	// add disconnect command
	shell.AddCmd(&ishell.Cmd{
		Name:     "disconnect",
		Help:     "disconnect from broker",
		LongHelp: `disconnect`,
		Func: func(ctx *ishell.Context) {
			// check state
			if c == nil {
				shell.Println("Failed: not connected")
				return
			}

			// check args
			if len(ctx.Args) != 0 {
				shell.Println("failed: missing arguments")
				return
			}

			// disconnect
			err := c.Disconnect(timeout)
			if err != nil {
				shell.Printf("Failed: %s\n", err.Error())
				c = nil
				return
			}

			// reset
			c = nil

			// print config
			shell.Println("Disconnected!")
		},
	})

	// run shell
	shell.Run()
}
