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
	"errors"
	"fmt"
	"sync"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/session"
	"github.com/gomqtt/transport"
	"github.com/satori/go.uuid"
	"gopkg.in/tomb.v2"
)

// ErrClientNotConnected may be returned by Publish and Close when the clients
// is not yet fully connected or has been already closed.
var ErrClientNotConnected = errors.New("client not connected")

const (
	clientConnecting byte = iota
	clientConnected
	clientDisconnected
)

// Client is a single client connected to the broker.
type Client struct {
	broker *Broker
	conn   transport.Conn

	Session session.Session // TODO: How do we get that session?
	UUID    string
	Info    interface{}

	state *state
	clean bool

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// NewClient takes over responsibility over a connection and returns a Client.
func NewClient(broker *Broker, conn transport.Conn) *Client {
	c := &Client{
		broker: broker,
		conn:   conn,
		UUID:   uuid.NewV1().String(),
		state:  newState(clientConnecting),
	}

	c.tomb.Go(c.processor)

	return c
}

// Publish will send a PublishPacket to the client and initiate QOS flows.
func (c *Client) Publish(topic string, payload []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check state
	if c.state.get() != clientConnected {
		return ErrClientNotConnected
	}

	// TODO: read QOS from cached subscriptions

	publish := packet.NewPublishPacket()
	publish.Topic = []byte(topic)
	publish.Payload = payload

	// store packet if at least qos 1
	if publish.QOS > 0 {
		err := c.Session.SavePacket(session.Outgoing, publish)
		if err != nil {
			return c.cleanup(err, true)
		}
	}

	// send packet
	err := c.send(publish)
	if err != nil {
		return c.cleanup(err, false)
	}

	return nil
}

// Close will cleanly close the connected client.
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() < clientConnecting {
		return ErrClientNotConnected
	}

	// close connection
	err := c.cleanup(nil, true)

	// shutdown goroutines
	c.tomb.Kill(nil)

	// wait for all goroutines to exit
	// goroutines will send eventual errors through the callback
	c.tomb.Wait()

	// do cleanup
	return err
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	c.log("%s - New Connection", c.UUID)

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			if c.state.get() == clientDisconnected {
				return c.die(nil, false)
			}

			// TODO: If the connection has been dropped send the will message.

			// die on any other error
			return c.die(err, false)
		}

		c.log("%s - Received: %s", c.UUID, pkt.String())

		// TODO: Handle errors

		switch pkt.Type() {
		case packet.CONNECT:
			err = c.processConnect(pkt.(*packet.ConnectPacket))
		case packet.SUBSCRIBE:
			err = c.processSubscribe(pkt.(*packet.SubscribePacket))
		case packet.UNSUBSCRIBE:
			err = c.processUnsubscribe(pkt.(*packet.UnsubscribePacket))
		case packet.PUBLISH:
			err = c.processPublish(pkt.(*packet.PublishPacket))
		case packet.PUBACK:
			err = c.processPubackAndPubcomp(pkt.(*packet.PubackPacket).PacketID)
		case packet.PUBCOMP:
			err = c.processPubackAndPubcomp(pkt.(*packet.PubcompPacket).PacketID)
		case packet.PUBREC:
			err = c.processPubrec(pkt.(*packet.PubrecPacket).PacketID)
		case packet.PUBREL:
			err = c.processPubrel(pkt.(*packet.PubrelPacket).PacketID)
		case packet.PINGREQ:
			err = c.processPingreq()
		case packet.DISCONNECT:
			err = c.processDisconnect()
		}

		// return eventual error
		if err != nil {
			return err // error has already been cleaned
		}
	}
}

// handle an incoming ConnackPacket
func (c *Client) processConnect(pkt *packet.ConnectPacket) error {
	connack := packet.NewConnackPacket()
	connack.ReturnCode = packet.ConnectionAccepted
	connack.SessionPresent = false

	// TODO: authenticate client
	// TODO: retrieve session
	c.state.set(clientConnected)

	err := c.send(connack)
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming PingreqPacket
func (c *Client) processPingreq() error {
	err := c.send(packet.NewPingrespPacket())
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming SubscribePacket
func (c *Client) processSubscribe(pkt *packet.SubscribePacket) error {
	suback := packet.NewSubackPacket()
	suback.ReturnCodes = make([]byte, 0)
	suback.PacketID = pkt.PacketID

	for _, subscription := range pkt.Subscriptions {
		// TODO: properly granted qos
		c.broker.QueueBackend.Subscribe(c, string(subscription.Topic))
		suback.ReturnCodes = append(suback.ReturnCodes, subscription.QOS)
	}

	err := c.send(suback)
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming UnsubscribePacket
func (c *Client) processUnsubscribe(pkt *packet.UnsubscribePacket) error {
	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = pkt.PacketID

	for _, topic := range pkt.Topics {
		c.broker.QueueBackend.Unsubscribe(c, string(topic))
	}

	err := c.send(unsuback)
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming PublishPacket
func (c *Client) processPublish(publish *packet.PublishPacket) error {
	if publish.QOS == 1 {
		puback := packet.NewPubackPacket()
		puback.PacketID = publish.PacketID

		// acknowledge qos 1 publish
		err := c.send(puback)
		if err != nil {
			return c.die(err, false)
		}
	}

	if publish.QOS == 2 {
		// store packet
		err := c.Session.SavePacket(session.Incoming, publish)
		if err != nil {
			return c.die(err, true)
		}

		pubrec := packet.NewPubrecPacket()
		pubrec.PacketID = publish.PacketID

		// signal qos 2 publish
		err = c.send(pubrec)
		if err != nil {
			return c.die(err, false)
		}
	}

	if publish.QOS <= 1 {
		// publish packet to others
		c.broker.QueueBackend.Publish(c, string(publish.Topic), publish.Payload)
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) processPubackAndPubcomp(packetID uint16) error {
	// remove packet from store
	c.Session.DeletePacket(session.Outgoing, packetID)

	return nil
}

// handle an incoming PubrecPacket
func (c *Client) processPubrec(packetID uint16) error {
	// allocate packet
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	// overwrite stored PublishPacket with PubrelPacket
	err := c.Session.SavePacket(session.Outgoing, pubrel)
	if err != nil {
		return c.die(err, true)
	}

	// send packet
	err = c.send(pubrel)
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming PubrelPacket
func (c *Client) processPubrel(packetID uint16) error {
	// get packet from store
	pkt, err := c.Session.LookupPacket(session.Incoming, packetID)
	if err != nil {
		return c.die(err, true)
	}

	// get packet from store
	publish, ok := pkt.(*packet.PublishPacket)
	if !ok {
		return nil // ignore a wrongly sent PubrelPacket
	}

	pubcomp := packet.NewPubcompPacket()
	pubcomp.PacketID = publish.PacketID

	// acknowledge PublishPacket
	err = c.send(pubcomp)
	if err != nil {
		return c.die(err, false)
	}

	// remove packet from store
	err = c.Session.DeletePacket(session.Incoming, packetID)
	if err != nil {
		return c.die(err, true)
	}

	// publish packet to others
	c.broker.QueueBackend.Publish(c, string(publish.Topic), publish.Payload)

	return nil
}

// handle an incoming DisconnectPacket
func (c *Client) processDisconnect() error {
	// mark client as cleanly disconnected
	c.state.set(clientDisconnected)

	return nil
}

/* helpers */

// will try to cleanup as many resources as possible
func (c *Client) cleanup(err error, close bool) error {
	// remove client from the queue
	c.broker.QueueBackend.Remove(c)

	// ensure that the connection gets closed
	if close {
		_err := c.conn.Close()
		if err == nil {
			err = _err
		}
	}

	// reset store
	if c.clean {
		_err := c.Session.Reset()
		if err == nil {
			err = _err
		}
	}

	return err
}

// used for closing and cleaning up from inside internal goroutines
func (c *Client) die(err error, close bool) error {
	c.finish.Do(func() {
		err = c.cleanup(err, close)

		// report error
		if err != nil {
			// TODO: what happens with an internal error?
			c.log("%s - Internal Error: %s", c.UUID, err)
		}
	})

	return err
}

// sends packet
func (c *Client) send(pkt packet.Packet) error {
	err := c.conn.Send(pkt)
	if err != nil {
		return err
	}

	c.log("%s - Sent: %s", c.UUID, pkt.String())

	return nil
}

// log a message
func (c *Client) log(format string, a ...interface{}) {
	if c.broker.Logger != nil {
		c.broker.Logger(fmt.Sprintf(format, a...))
	}
}
