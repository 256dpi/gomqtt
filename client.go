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
	"sync"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"github.com/satori/go.uuid"
	"gopkg.in/tomb.v2"
)

// A Client represents a remote client that is connected to the broker.
type Client interface {
	// Publish will send a Message to the client and initiate QOS flows.
	Publish(msg *packet.Message) bool

	// Close will immediately close the connection. When clean=true the client
	// will be marked as cleanly disconnected, and the will messages will not
	// get dispatched.
	Close(clean bool)

	// Context returns the associated context.
	Context() *Context
}

const (
	clientConnecting byte = iota
	clientConnected
	clientDisconnected
)

// ErrExpectedConnect is returned when the first received packet is not a
// ConnectPacket.
var ErrExpectedConnect = errors.New("expected ConnectPacket")

type remoteClient struct {
	broker *Broker
	conn   transport.Conn

	session Session
	context *Context

	out   chan *packet.Message
	state *state

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// newRemoteClient takes over a connection and returns a remoteClient
func newRemoteClient(broker *Broker, conn transport.Conn) *remoteClient {
	c := &remoteClient{
		broker:  broker,
		conn:    conn,
		context: NewContext(),
		out:     make(chan *packet.Message),
		state:   newState(clientConnecting),
	}

	c.Context().Set("uuid", uuid.NewV1().String())

	// start processor
	c.tomb.Go(c.processor)

	return c
}

// Context returns the associated context. Every client will already have the
// "uuid" value set in the context.
func (c *remoteClient) Context() *Context {
	return c.context
}

// Publish will send a Message to the client and initiate QOS flows.
func (c *remoteClient) Publish(msg *packet.Message) bool {
	select {
	case c.out <- msg:
		return true
	case <-c.tomb.Dying():
		return false
	}
}

// Close will immediately close the connection.
func (c *remoteClient) Close(clean bool) {
	if clean {
		// mark client as cleanly disconnected
		c.state.set(clientDisconnected)
	}

	// close underlying connection (triggers cleanup)
	c.conn.Close()
}

/* processor goroutine */

// processes incoming packets
func (c *remoteClient) processor() error {
	first := true

	c.log(NewConnectionLogEvent, c, nil, nil)

	// set initial read timeout
	c.conn.SetReadTimeout(c.broker.ConnectTimeout)

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

		c.log(PacketReceivedLogEvent, c, pkt, nil)

		if first {
			// get connect
			connect, ok := pkt.(*packet.ConnectPacket)
			if !ok {
				return c.die(ErrExpectedConnect, true)
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
func (c *remoteClient) processConnect(pkt *packet.ConnectPacket) error {
	connack := packet.NewConnackPacket()
	connack.ReturnCode = packet.ConnectionAccepted
	connack.SessionPresent = false

	// authenticate
	ok, err := c.broker.Backend.Authenticate(c, pkt.Username, pkt.Password)
	if err != nil {
		c.die(err, true)
	}

	// check authentication
	if !ok {
		// set state
		c.state.set(clientDisconnected)

		// set return code
		connack.ReturnCode = packet.ErrNotAuthorized

		// send connack
		err = c.send(connack)
		if err != nil {
			return c.die(err, false)
		}

		// close client
		c.die(nil, true)
	}

	// set state
	c.state.set(clientConnected)

	// set keep alive
	if pkt.KeepAlive > 0 {
		c.conn.SetReadTimeout(time.Duration(pkt.KeepAlive) * 1500 * time.Millisecond)
	} else {
		c.conn.SetReadTimeout(0)
	}

	// retrieve session
	sess, resumed, err := c.broker.Backend.Setup(c, pkt.ClientID, pkt.CleanSession)
	if err != nil {
		return c.die(err, true)
	}

	// set session present
	connack.SessionPresent = !pkt.CleanSession && resumed

	// assign session
	c.session = sess

	// save will if present
	if pkt.Will != nil {
		err = c.session.SaveWill(pkt.Will)
		if err != nil {
			return c.die(err, true)
		}
	}

	// send connack
	err = c.send(connack)
	if err != nil {
		return c.die(err, false)
	}

	// start sender
	c.tomb.Go(c.sender)

	// retrieve stored packets
	packets, err := c.session.AllPackets(outgoing)
	if err != nil {
		return c.die(err, true)
	}

	// resend stored packets
	for _, pkt := range packets {
		publish, ok := pkt.(*packet.PublishPacket)
		if ok {
			// set the dup flag on a publish packet
			publish.Dup = true
		}

		err = c.send(pkt)
		if err != nil {
			return c.die(err, false)
		}
	}

	// attempt to restore client if not clean
	if !pkt.CleanSession {
		err = c.broker.Backend.Restore(c)
		if err != nil {
			return c.die(err, true)
		}
	}

	return nil
}

// handle an incoming PingreqPacket
func (c *remoteClient) processPingreq() error {
	err := c.send(packet.NewPingrespPacket())
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming SubscribePacket
func (c *remoteClient) processSubscribe(pkt *packet.SubscribePacket) error {
	suback := packet.NewSubackPacket()
	suback.ReturnCodes = make([]byte, len(pkt.Subscriptions))
	suback.PacketID = pkt.PacketID

	// handle contained subscriptions
	for i, subscription := range pkt.Subscriptions {
		// save subscription in session
		err := c.session.SaveSubscription(&subscription)
		if err != nil {
			return c.die(err, true)
		}

		// subscribe client to queue
		err = c.broker.Backend.Subscribe(c, subscription.Topic)
		if err != nil {
			return c.die(err, true)
		}

		// save granted qos
		suback.ReturnCodes[i] = subscription.QOS
	}

	// send suback
	err := c.send(suback)
	if err != nil {
		return c.die(err, false)
	}

	// queue retained messages
	for _, sub := range pkt.Subscriptions {
		err := c.broker.Backend.QueueRetained(c, sub.Topic)
		if err != nil {
			return c.die(err, true)
		}
	}

	return nil
}

// handle an incoming UnsubscribePacket
func (c *remoteClient) processUnsubscribe(pkt *packet.UnsubscribePacket) error {
	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = pkt.PacketID

	for _, topic := range pkt.Topics {
		// unsubscribe client from queue
		err := c.broker.Backend.Unsubscribe(c, topic)
		if err != nil {
			return c.die(err, true)
		}

		// remove subscription from session
		err = c.session.DeleteSubscription(topic)
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
func (c *remoteClient) processPublish(publish *packet.PublishPacket) error {
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
		err := c.session.SavePacket(incoming, publish)
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
		err := c.finishPublish(&publish.Message)
		if err != nil {
			return c.die(err, true)
		}
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *remoteClient) processPubackAndPubcomp(packetID uint16) error {
	// remove packet from store
	c.session.DeletePacket(outgoing, packetID)

	return nil
}

// handle an incoming PubrecPacket
func (c *remoteClient) processPubrec(packetID uint16) error {
	// allocate packet
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	// overwrite stored PublishPacket with PubrelPacket
	err := c.session.SavePacket(outgoing, pubrel)
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
func (c *remoteClient) processPubrel(packetID uint16) error {
	// get packet from store
	pkt, err := c.session.LookupPacket(incoming, packetID)
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
	err = c.session.DeletePacket(incoming, packetID)
	if err != nil {
		return c.die(err, true)
	}

	// publish packet to others
	err = c.finishPublish(&publish.Message)
	if err != nil {
		return c.die(err, true)
	}

	return nil
}

// handle an incoming DisconnectPacket
func (c *remoteClient) processDisconnect() error {
	// mark client as cleanly disconnected
	c.state.set(clientDisconnected)

	// clear will
	err := c.session.ClearWill()
	if err != nil {
		return c.die(err, true)
	}

	return nil
}

/* sender goroutine */

// sends outgoing messages
func (c *remoteClient) sender() error {
	for {
		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		case msg := <-c.out:
			publish := packet.NewPublishPacket()
			publish.Message = *msg

			// get stored subscription
			sub, err := c.session.LookupSubscription(publish.Message.Topic)
			if err != nil {
				return c.die(err, true)
			}
			if sub != nil {
				// respect maximum qos
				if publish.Message.QOS > sub.QOS {
					publish.Message.QOS = sub.QOS
				}
			}

			// set packet id
			if publish.Message.QOS > 0 {
				publish.PacketID = c.session.PacketID()
			}

			// store packet if at least qos 1
			if publish.Message.QOS > 0 {
				err := c.session.SavePacket(outgoing, publish)
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

func (c *remoteClient) finishPublish(msg *packet.Message) error {
	// check retain flag
	if msg.Retain {
		if len(msg.Payload) > 0 {
			err := c.broker.Backend.StoreRetained(c, msg)
			if err != nil {
				return err
			}
		} else {
			err := c.broker.Backend.ClearRetained(c, msg.Topic)
			if err != nil {
				return err
			}
		}
	}

	// reset an existing retain flag
	msg.Retain = false

	// publish message to others
	return c.broker.Backend.Publish(c, msg)
}

// will try to cleanup as many resources as possible
func (c *remoteClient) cleanup(err error, close bool) error {
	// check session
	if c.session != nil && c.state.get() != clientDisconnected {
		// get will
		will, _err := c.session.LookupWill()
		if err == nil {
			err = _err
		}

		// publish will message
		if will != nil {
			_err = c.finishPublish(will)
			if err == nil {
				err = _err
			}
		}
	}

	// remove client from the queue
	_err := c.broker.Backend.Terminate(c)
	if err == nil {
		err = _err
	}

	// ensure that the connection gets closed
	if close {
		_err := c.conn.Close()
		if err == nil {
			err = _err
		}
	}

	c.log(LostConnectionLogEvent, c, nil, nil)

	return err
}

// used for closing and cleaning up from inside internal goroutines
func (c *remoteClient) die(err error, close bool) error {
	c.finish.Do(func() {
		err = c.cleanup(err, close)

		// report error
		if err != nil {
			c.log(ErrorLogEvent, c, nil, err)
		}
	})

	return err
}

// sends packet
func (c *remoteClient) send(pkt packet.Packet) error {
	err := c.conn.Send(pkt)
	if err != nil {
		return err
	}

	c.log(PacketSentLogEvent, c, pkt, nil)

	return nil
}

// log a message
func (c *remoteClient) log(event LogEvent, client Client, pkt packet.Packet, err error) {
	if c.broker.Logger != nil {
		c.broker.Logger(event, client, pkt, err)
	}
}
