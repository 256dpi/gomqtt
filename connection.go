// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broker

import (
	"sync"
	"fmt"

	"github.com/gomqtt/stream"
	"github.com/gomqtt/packet"
	"github.com/satori/go.uuid"
)

type Connection struct {
	broker *Broker
	stream stream.Stream

	uuid string
	quit chan struct{}
	start sync.WaitGroup
	finish sync.WaitGroup
}

func NewConnection(broker *Broker, stream stream.Stream) *Connection {
	c := &Connection{
		broker: broker,
		stream: stream,
		uuid: uuid.NewV1().String(),
		quit: make(chan struct{}),
	}

	c.start.Add(1)
	c.finish.Add(1)
	go c.process()

	return c
}

func (c *Connection) Close() {
	close(c.quit)

	c.finish.Wait()
}

func (c *Connection) process() {
	c.start.Done()
	defer c.finish.Done()

	fmt.Println("new connection: " + c.uuid)

	for {
		select {
		case <-c.quit:
			return
		case msg, ok := <-c.stream.Incoming():
			if !ok {
				fmt.Println("lost connection: " + c.uuid)

				err := c.stream.Error()
				if err != nil {
					fmt.Println(err)
				}

				c.broker.queueBackend.Remove(c)

				return
			}

			switch msg.Type() {
			case packet.CONNECT:
				fmt.Println("received connect")
				_, ok := msg.(*packet.ConnectPacket)

				if ok {
					ca := packet.NewConnackPacket()
					ca.ReturnCode = packet.ConnectionAccepted
					ca.SessionPresent = false
					c.stream.Send(ca)
				}
			case packet.PINGREQ:
				_, ok := msg.(*packet.PingrespPacket)

				if ok {
					c.stream.Send(packet.NewPingrespPacket())
				}
			case packet.PUBLISH:
				pp, ok := msg.(*packet.PublishPacket)

				if ok {
					c.broker.queueBackend.Publish(pp)
				}
			case packet.SUBSCRIBE:
				sp, ok := msg.(*packet.SubscribePacket)

				if ok {
					m := packet.NewSubackPacket()
					m.ReturnCodes = make([]byte, 0)
					m.PacketID = sp.PacketID

					for _, s := range sp.Subscriptions {
						c.broker.queueBackend.Subscribe(c, string(s.Topic))
						m.ReturnCodes = append(m.ReturnCodes, s.QOS)
					}

					c.stream.Send(m)
				}
			}
		}
	}
}
