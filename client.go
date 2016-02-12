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
	"github.com/gomqtt/session"
)

type Client struct {
	broker *Broker
	conn   transport.Conn

	Session session.Session // TODO: How do we get that session?
	UUID string
	Info interface{}

	tomb tomb.Tomb
}

func NewClient(broker *Broker, conn transport.Conn) *Client {
	c := &Client{
		broker: broker,
		conn: conn,
		UUID: uuid.NewV1().String(),
	}

	c.tomb.Go(c.processor)

	return c
}

func (c *Client) Publish(pkt *packet.PublishPacket) {
	c.send(pkt) // TODO: handle errors
}

func (c *Client) Close() {
	c.tomb.Kill(nil)
	c.tomb.Wait()
}

/* processor goroutine */

func (c *Client) processor() error {
	c.log("%s - New Connection", c.UUID)

	for {
		pkt, err := c.conn.Receive()
		if err != nil {
			c.broker.QueueBackend.Remove(c)

			c.log("%s - Error: %s", c.UUID, err)
			return err
		}

		c.log("%s - Received: %s", c.UUID, pkt.String())

		// TODO: Handle errors

		switch pkt.Type() {
		case packet.CONNECT:
			c.processConnect(pkt.(*packet.ConnectPacket))
		case packet.PINGREQ:
			c.processPingreq(pkt.(*packet.PingreqPacket))
		case packet.SUBSCRIBE:
			c.processSubscribe(pkt.(*packet.SubscribePacket))
		case packet.UNSUBSCRIBE:
			c.processUnsubscribe(pkt.(*packet.UnsubscribePacket))
		case packet.PUBLISH:
			c.processPublish(pkt.(*packet.PublishPacket))
		case packet.PUBACK:
			//err = c.processPubackAndPubcomp(pkt.(*packet.PubackPacket).PacketID)
		case packet.PUBCOMP:
			//err = c.processPubackAndPubcomp(pkt.(*packet.PubcompPacket).PacketID)
		case packet.PUBREC:
			//err = c.processPubrec(pkt.(*packet.PubrecPacket).PacketID)
		case packet.PUBREL:
			//err = c.processPubrel(pkt.(*packet.PubrelPacket).PacketID)
		}
	}
}

func (c *Client) processConnect(pkt *packet.ConnectPacket) error {
	connack := packet.NewConnackPacket()
	connack.ReturnCode = packet.ConnectionAccepted
	connack.SessionPresent = false
	return c.send(connack)
}

func (c *Client) processPingreq(pkt *packet.PingreqPacket) error {
	return c.send(packet.NewPingrespPacket())
}

func (c *Client) processPublish(pkt *packet.PublishPacket) error {
	c.broker.QueueBackend.Publish(pkt)
	return nil
}

func (c *Client) processSubscribe(pkt *packet.SubscribePacket) error {
	suback := packet.NewSubackPacket()
	suback.ReturnCodes = make([]byte, 0)
	suback.PacketID = pkt.PacketID

	for _, subscription := range pkt.Subscriptions {
		c.broker.QueueBackend.Subscribe(c, string(subscription.Topic))
		suback.ReturnCodes = append(suback.ReturnCodes, subscription.QOS)
	}

	return c.send(suback)
}

func (c *Client) processUnsubscribe(pkt *packet.UnsubscribePacket) error {
	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = pkt.PacketID

	for _, topic := range pkt.Topics {
		c.broker.QueueBackend.Unsubscribe(c, string(topic))
	}

	return c.send(unsuback)
}

/* helpers */

func (c *Client) send(pkt packet.Packet) error {
	c.log("%s - Sent: %s", c.UUID, pkt.String())
	return c.conn.Send(pkt)
}

// log a message
func (c *Client) log(format string, a ...interface{}) {
	if c.broker.Logger != nil {
		c.broker.Logger(fmt.Sprintf(format, a...))
	}
}
