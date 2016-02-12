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
	info interface{}

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

func (c *Client) publish(pkt *packet.PublishPacket) {
	c.send(pkt) // TODO: handle errors
}

func (c *Client) send(pkt packet.Packet) error {
	fmt.Printf("%s - Sent: %s\n", c.uuid, pkt.String())
	return c.conn.Send(pkt)
}

func (c *Client) process() error {
	fmt.Printf("%s - New Connection\n", c.uuid)

	for {
		pkt, err := c.conn.Receive()
		if err != nil {
			c.broker.queueBackend.Remove(c)

			fmt.Printf("%s - Error: %s\n", c.uuid, err)
			return err
		}

		fmt.Printf("%s - Received: %s\n", c.uuid, pkt.String())

		// TODO: Handle errors

		switch pkt.Type() {
		case packet.CONNECT:
			c.handleConnect(pkt.(*packet.ConnectPacket))
		case packet.PINGREQ:
			c.handlePingreq(pkt.(*packet.PingreqPacket))
		case packet.PUBLISH:
			c.handlePublish(pkt.(*packet.PublishPacket))
		case packet.SUBSCRIBE:
			c.handleSubscribe(pkt.(*packet.SubscribePacket))
		}
	}
}

func (c *Client) handleConnect(pkt *packet.ConnectPacket) error {
	connack := packet.NewConnackPacket()
	connack.ReturnCode = packet.ConnectionAccepted
	connack.SessionPresent = false
	return c.send(connack)
}

func (c *Client) handlePingreq(pkt *packet.PingreqPacket) error {
	return c.send(packet.NewPingrespPacket())
}

func (c *Client) handlePublish(pkt *packet.PublishPacket) error {
	c.broker.queueBackend.Publish(pkt)
	return nil
}

func (c *Client) handleSubscribe(pkt *packet.SubscribePacket) error {
	suback := packet.NewSubackPacket()
	suback.ReturnCodes = make([]byte, 0)
	suback.PacketID = pkt.PacketID

	for _, subscription := range pkt.Subscriptions {
		c.broker.queueBackend.Subscribe(c, string(subscription.Topic))
		suback.ReturnCodes = append(suback.ReturnCodes, subscription.QOS)
	}

	return c.send(suback)
}

func (c *Client) close() {
	c.tomb.Kill(nil)
	c.tomb.Wait()
}
