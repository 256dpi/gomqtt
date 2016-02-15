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
	"sync"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/session"
	"github.com/gomqtt/transport"
	"github.com/satori/go.uuid"
	"gopkg.in/tomb.v2"
)

const (
	clientConnecting byte = iota
	clientConnected
	clientDisconnected
)

// Client is a single client connected to the broker.
type Client struct {
	broker *Broker
	conn   transport.Conn

	Session session.Session
	UUID    string
	Info    interface{}

	out   chan *packet.Message
	state *state
	clean bool

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// NewClient takes over responsibility over a connection and returns a Client.
func NewClient(broker *Broker, conn transport.Conn) *Client {
	c := &Client{
		broker:  broker,
		conn:    conn,
		out:     make(chan *packet.Message),
		UUID:    uuid.NewV1().String(),
		Session: session.NewMemorySession(),
		state:   newState(clientConnecting),
	}

	c.tomb.Go(c.processor)
	c.tomb.Go(c.sender)

	return c
}

// Publish will send a Message to the client and initiate QOS flows.
func (c *Client) Publish(msg *packet.Message) bool {
	select {
	case c.out <- msg:
		return true
	case <-c.tomb.Dying():
		return false
	}
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	first := true

	c.log("%s - New Connection", c.UUID)

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			if c.state.get() == clientDisconnected {
				return c.die(nil, false)
			}

			// die on any other error
			return c.die(err, false)
		}

		c.log("%s - Received: %s", c.UUID, pkt.String())

		if first {
			// get connect
			connect, ok := pkt.(*packet.ConnectPacket)
			if !ok {
				return c.die(fmt.Errorf("expected connect"), true)
			}

			// process connect
			err = c.processConnect(connect)
			first = false
		}

		switch _pkt := pkt.(type) {
		case *packet.SubscribePacket:
			err = c.processSubscribe(_pkt)
		case *packet.UnsubscribePacket:
			err = c.processUnsubscribe(_pkt)
		case *packet.PublishPacket:
			err = c.processPublish(_pkt)
		case *packet.PubackPacket:
			err = c.processPubackAndPubcomp(_pkt.PacketID)
		case *packet.PubcompPacket:
			err = c.processPubackAndPubcomp(_pkt.PacketID)
		case *packet.PubrecPacket:
			err = c.processPubrec(_pkt.PacketID)
		case *packet.PubrelPacket:
			err = c.processPubrel(_pkt.PacketID)
		case *packet.PingreqPacket:
			err = c.processPingreq()
		case *packet.DisconnectPacket:
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
	c.state.set(clientConnected)

	// set clean flag
	c.clean = pkt.CleanSession

	if len(pkt.ClientID) > 0 {
		// retrieve session
		sess, err := c.broker.Backend.GetSession(c, pkt.ClientID)
		if err != nil {
			return c.die(err, true)
		}

		// reset session on clean
		if c.clean {
			err = sess.Reset()
			if err != nil {
				return c.die(err, true)
			}
		}

		// assign session
		c.Session = sess
	}

	// save will if present
	if pkt.Will != nil {
		c.Session.SaveWill(pkt.Will)
	}

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

	var retainedMessages []*packet.Message

	for _, subscription := range pkt.Subscriptions {
		// save subscription in session
		err := c.Session.SaveSubscription(&subscription)
		if err != nil {
			return c.die(err, true)
		}

		// subscribe client to queue
		msgs, err := c.broker.Backend.Subscribe(c, subscription.Topic)
		if err != nil {
			return c.die(err, true)
		}

		// cache retained messages
		retainedMessages = append(retainedMessages, msgs...)

		// save granted qos
		suback.ReturnCodes = append(suback.ReturnCodes, subscription.QOS)
	}

	// send suback
	err := c.send(suback)
	if err != nil {
		return c.die(err, false)
	}

	// send messages
	for _, msg := range retainedMessages {
		c.out <- msg
	}

	return nil
}

// handle an incoming UnsubscribePacket
func (c *Client) processUnsubscribe(pkt *packet.UnsubscribePacket) error {
	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = pkt.PacketID

	for _, topic := range pkt.Topics {
		// unsubscribe client from queue
		err := c.broker.Backend.Unsubscribe(c, topic)
		if err != nil {
			return c.die(err, true)
		}

		// remove subscription from session
		err = c.Session.DeleteSubscription(topic)
		if err != nil {
			return c.die(err, true)
		}
	}

	err := c.send(unsuback)
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming PublishPacket
func (c *Client) processPublish(publish *packet.PublishPacket) error {
	if publish.Message.QOS == 1 {
		puback := packet.NewPubackPacket()
		puback.PacketID = publish.PacketID

		// acknowledge qos 1 publish
		err := c.send(puback)
		if err != nil {
			return c.die(err, false)
		}
	}

	if publish.Message.QOS == 2 {
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

	if publish.Message.QOS <= 1 {
		// publish packet to others
		err := c.broker.Backend.Publish(c, &publish.Message)
		if err != nil {
			return c.die(err, true)
		}
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
	err = c.broker.Backend.Publish(c, &publish.Message)
	if err != nil {
		return c.die(err, true)
	}

	return nil
}

// handle an incoming DisconnectPacket
func (c *Client) processDisconnect() error {
	// mark client as cleanly disconnected
	c.state.set(clientDisconnected)

	// clear will
	err := c.Session.ClearWill()
	if err != nil {
		return c.die(err, true)
	}

	return nil
}

/* sender goroutine */

// sends outgoing messages
func (c *Client) sender() error {
	for {
		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		case msg := <-c.out:
			publish := packet.NewPublishPacket()
			publish.Message = *msg

			// get stored subscription
			sub, err := c.Session.LookupSubscription(publish.Message.Topic)
			if err != nil {
				return c.die(err, true)
			}

			// check subscription
			if sub == nil {
				return c.die(fmt.Errorf("subscription not found in session"), true)
			}

			// respect maximum qos
			if publish.Message.QOS > sub.QOS {
				publish.Message.QOS = sub.QOS
			}

			// set packet id
			if publish.Message.QOS > 0 {
				publish.PacketID = c.Session.PacketID()
			}

			// store packet if at least qos 1
			if publish.Message.QOS > 0 {
				err := c.Session.SavePacket(session.Outgoing, publish)
				if err != nil {
					return c.die(err, true)
				}
			}

			// send packet
			err = c.send(publish)
			if err != nil {
				return c.die(err, false)
			}
		}
	}
}

/* helpers */

// will try to cleanup as many resources as possible
func (c *Client) cleanup(err error, close bool) error {
	// remove client from the queue
	_err := c.broker.Backend.Remove(c)
	if err == nil {
		err = _err
	}

	// get will
	will, _err := c.Session.LookupWill()
	if err == nil {
		err = _err
	}

	// publish will message
	if will != nil {
		c.broker.Backend.Publish(c, will)
	}

	// ensure that the connection gets closed
	if close {
		_err := c.conn.Close()
		if err == nil {
			err = _err
		}
	}

	// reset session
	if c.clean {
		_err := c.Session.Reset()
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
