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
	"fmt"

	"github.com/gomqtt/transport"
	"github.com/gomqtt/packet"
	"github.com/satori/go.uuid"
	"gopkg.in/tomb.v2"
)

type Client struct {
	broker *Broker
	conn transport.Conn
	uuid string

	tomb tomb.Tomb
}

func NewClient(broker *Broker, conn transport.Conn) *Client {
	c := &Client{
		broker: broker,
		conn: conn,
		uuid: uuid.NewV1().String(),
	}

	c.tomb.Go(c.process)

	return c
}

func (c *Client) close() {
	c.tomb.Kill(nil)
	c.tomb.Wait()
}

func (c *Client) publish(pkt *packet.PublishPacket) {
	c.conn.Send(pkt)
}

func (c *Client) process() error {
	fmt.Println("new connection: " + c.uuid)

	for {
		pkt, err := c.conn.Receive()
		if err != nil {
			c.broker.queueBackend.Remove(c)

			fmt.Println(err)
			return err
		}

		switch pkt.Type() {
		case packet.CONNECT:
			fmt.Println("received connect")
			_, ok := pkt.(*packet.ConnectPacket)

			if ok {
				connack := packet.NewConnackPacket()
				connack.ReturnCode = packet.ConnectionAccepted
				connack.SessionPresent = false
				c.conn.Send(connack)
			}
		case packet.PINGREQ:
			_, ok := pkt.(*packet.PingrespPacket)

			if ok {
				c.conn.Send(packet.NewPingrespPacket())
			}
		case packet.PUBLISH:
			publish, ok := pkt.(*packet.PublishPacket)

			if ok {
				c.broker.queueBackend.Publish(publish)
			}
		case packet.SUBSCRIBE:
			subscribe, ok := pkt.(*packet.SubscribePacket)

			if ok {
				suback := packet.NewSubackPacket()
				suback.ReturnCodes = make([]byte, 0)
				suback.PacketID = subscribe.PacketID

				for _, subscription := range subscribe.Subscriptions {
					c.broker.queueBackend.Subscribe(c, string(subscription.Topic))
					suback.ReturnCodes = append(suback.ReturnCodes, subscription.QOS)
				}

				c.conn.Send(suback)
			}
		}
	}
}
