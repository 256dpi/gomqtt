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
	"net"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"gopkg.in/tomb.v2"
)

const (
	clientConnecting byte = iota
	clientConnected
	clientDisconnected
)

// ErrExpectedConnect is returned when the first received packet is not a
// ConnectPacket.
var ErrExpectedConnect = errors.New("expected a ConnectPacket as the first packet")

// A Client represents a remote client that is connected to the broker.
type Client struct {
	Context *Context

	engine *Engine
	conn   transport.Conn

	clientID     string
	cleanSession bool
	session      Session

	out   chan *packet.Message
	state *state

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// newClient takes over a connection and returns a Client
func newClient(engine *Engine, conn transport.Conn) *Client {
	c := &Client{
		Context: NewContext(),
		engine:  engine,
		conn:    conn,
		out:     make(chan *packet.Message),
		state:   newState(clientConnecting),
	}

	// start processor
	c.tomb.Go(c.processor)

	return c
}

// Session returns the current Session used by the client.
func (c *Client) Session() Session {
	return c.session
}

// CleanSession returns whether the client requested a clean session during connect.
func (c *Client) CleanSession() bool {
	return c.cleanSession
}

// ClientID returns the supplied client id during connect.
func (c *Client) ClientID() string {
	return c.clientID
}

// RemoteAddr returns the client's remote net address from the
// underlying connection.
func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
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

// Close will immediately close the connection. When clean=true the client
// will be marked as cleanly disconnected, and the will message will not
// get dispatched.
func (c *Client) Close(clean bool) {
	if clean {
		// mark client as cleanly disconnected
		c.state.set(clientDisconnected)
	}

	// close underlying connection (triggers cleanup)
	c.conn.Close()
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	first := true

	c.log(NewConnection, c, nil, nil, nil)

	// set initial read timeout
	c.conn.SetReadTimeout(c.engine.ConnectTimeout)

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			return c.die(TransportError, err, false)
		}

		c.log(PacketReceived, c, pkt, nil, nil)

		if first {
			// get connect
			connect, ok := pkt.(*packet.ConnectPacket)
			if !ok {
				return c.die(ClientError, ErrExpectedConnect, true)
			}

			// process connect
			err = c.processConnect(connect)
			first = false
		}

		switch typedPkt := pkt.(type) {
		case *packet.SubscribePacket:
			err = c.processSubscribe(typedPkt)
		case *packet.UnsubscribePacket:
			err = c.processUnsubscribe(typedPkt)
		case *packet.PublishPacket:
			err = c.processPublish(typedPkt)
		case *packet.PubackPacket:
			err = c.processPubackAndPubcomp(typedPkt.PacketID)
		case *packet.PubcompPacket:
			err = c.processPubackAndPubcomp(typedPkt.PacketID)
		case *packet.PubrecPacket:
			err = c.processPubrec(typedPkt.PacketID)
		case *packet.PubrelPacket:
			err = c.processPubrel(typedPkt.PacketID)
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

	// authenticate
	ok, err := c.engine.Backend.Authenticate(c, pkt.Username, pkt.Password)
	if err != nil {
		c.die(BackendError, err, true)
	}

	// check authentication
	if !ok {
		// set return code
		connack.ReturnCode = packet.ErrNotAuthorized

		// send connack
		err = c.send(connack, false)
		if err != nil {
			return c.die(TransportError, err, false)
		}

		// close client
		return c.die(ClientError, nil, true)
	}

	// set state
	c.state.set(clientConnected)

	// add client to the brokers list
	c.engine.add(c)

	// set keep alive
	if pkt.KeepAlive > 0 {
		c.conn.SetReadTimeout(time.Duration(pkt.KeepAlive) * 1500 * time.Millisecond)
	} else {
		c.conn.SetReadTimeout(0)
	}

	// set values
	c.cleanSession = pkt.CleanSession
	c.clientID = pkt.ClientID

	// retrieve session
	s, resumed, err := c.engine.Backend.Setup(c, pkt.ClientID)
	if err != nil {
		return c.die(BackendError, err, true)
	}

	// reset the session if clean is requested
	if pkt.CleanSession {
		s.Reset()
	}

	// set session present
	connack.SessionPresent = !pkt.CleanSession && resumed

	// assign session
	c.session = s

	// save will if present
	if pkt.Will != nil {
		err = c.session.SaveWill(pkt.Will)
		if err != nil {
			return c.die(SessionError, err, true)
		}
	}

	// send connack
	err = c.send(connack, false)
	if err != nil {
		return c.die(TransportError, err, false)
	}

	// start sender
	c.tomb.Go(c.sender)

	// retrieve stored packets
	packets, err := c.session.AllPackets(outgoing)
	if err != nil {
		return c.die(SessionError, err, true)
	}

	// resend stored packets
	for _, pkt := range packets {
		publish, ok := pkt.(*packet.PublishPacket)
		if ok {
			// set the dup flag on a publish packet
			publish.Dup = true
		}

		err = c.send(pkt, true)
		if err != nil {
			return c.die(TransportError, err, false)
		}
	}

	// attempt to restore client if not clean
	if !pkt.CleanSession {
		// get stored subscriptions
		subs, err := s.AllSubscriptions()
		if err != nil {
			return c.die(SessionError, err, true)
		}

		// resubscribe subscriptions
		for _, sub := range subs {
			err = c.engine.Backend.Subscribe(c, sub.Topic)
			if err != nil {
				return c.die(BackendError, err, true)
			}
		}

		// begin with queueing offline messages
		err = c.engine.Backend.QueueOffline(c)
		if err != nil {
			return c.die(BackendError, err, true)
		}
	}

	return nil
}

// handle an incoming PingreqPacket
func (c *Client) processPingreq() error {
	err := c.send(packet.NewPingrespPacket(), true)
	if err != nil {
		return c.die(TransportError, err, false)
	}

	return nil
}

// handle an incoming SubscribePacket
func (c *Client) processSubscribe(pkt *packet.SubscribePacket) error {
	suback := packet.NewSubackPacket()
	suback.ReturnCodes = make([]byte, len(pkt.Subscriptions))
	suback.PacketID = pkt.PacketID

	// handle contained subscriptions
	for i, subscription := range pkt.Subscriptions {
		// save subscription in session
		err := c.session.SaveSubscription(&subscription)
		if err != nil {
			return c.die(SessionError, err, true)
		}

		// subscribe client to queue
		err = c.engine.Backend.Subscribe(c, subscription.Topic)
		if err != nil {
			return c.die(BackendError, err, true)
		}

		// save granted qos
		suback.ReturnCodes[i] = subscription.QOS
	}

	// send suback
	err := c.send(suback, true)
	if err != nil {
		return c.die(TransportError, err, false)
	}

	// queue retained messages
	for _, sub := range pkt.Subscriptions {
		err := c.engine.Backend.QueueRetained(c, sub.Topic)
		if err != nil {
			return c.die(BackendError, err, true)
		}
	}

	return nil
}

// handle an incoming UnsubscribePacket
func (c *Client) processUnsubscribe(pkt *packet.UnsubscribePacket) error {
	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = pkt.PacketID

	for _, topic := range pkt.Topics {
		// unsubscribe client from queue
		err := c.engine.Backend.Unsubscribe(c, topic)
		if err != nil {
			return c.die(BackendError, err, true)
		}

		// remove subscription from session
		err = c.session.DeleteSubscription(topic)
		if err != nil {
			return c.die(SessionError, err, true)
		}
	}

	err := c.send(unsuback, true)
	if err != nil {
		return c.die(TransportError, err, false)
	}

	return nil
}

// handle an incoming PublishPacket
func (c *Client) processPublish(publish *packet.PublishPacket) error {
	// handle unacknowledged and directly acknowledged messages
	if publish.Message.QOS <= 1 {
		err := c.handleMessage(&publish.Message)
		if err != nil {
			return c.die(BackendError, err, true)
		}
	}

	// handle qos 1 flow
	if publish.Message.QOS == 1 {
		puback := packet.NewPubackPacket()
		puback.PacketID = publish.PacketID

		// acknowledge qos 1 publish
		err := c.send(puback, true)
		if err != nil {
			return c.die(TransportError, err, false)
		}
	}

	// handle qos 2 flow
	if publish.Message.QOS == 2 {
		// store packet
		err := c.session.SavePacket(incoming, publish)
		if err != nil {
			return c.die(SessionError, err, true)
		}

		pubrec := packet.NewPubrecPacket()
		pubrec.PacketID = publish.PacketID

		// signal qos 2 publish
		err = c.send(pubrec, true)
		if err != nil {
			return c.die(TransportError, err, false)
		}
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) processPubackAndPubcomp(packetID uint16) error {
	// remove packet from store
	c.session.DeletePacket(outgoing, packetID)

	return nil
}

// handle an incoming PubrecPacket
func (c *Client) processPubrec(packetID uint16) error {
	// allocate packet
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	// overwrite stored PublishPacket with PubrelPacket
	err := c.session.SavePacket(outgoing, pubrel)
	if err != nil {
		return c.die(SessionError, err, true)
	}

	// send packet
	err = c.send(pubrel, true)
	if err != nil {
		return c.die(TransportError, err, false)
	}

	return nil
}

// handle an incoming PubrelPacket
func (c *Client) processPubrel(packetID uint16) error {
	// get packet from store
	pkt, err := c.session.LookupPacket(incoming, packetID)
	if err != nil {
		return c.die(SessionError, err, true)
	}

	// get packet from store
	publish, ok := pkt.(*packet.PublishPacket)
	if !ok {
		return nil // ignore a wrongly sent PubrelPacket
	}

	// publish packet to others
	err = c.handleMessage(&publish.Message)
	if err != nil {
		return c.die(BackendError, err, true)
	}

	// prepare pubcomp packet
	pubcomp := packet.NewPubcompPacket()
	pubcomp.PacketID = publish.PacketID

	// acknowledge PublishPacket
	err = c.send(pubcomp, true)
	if err != nil {
		return c.die(TransportError, err, false)
	}

	// remove packet from store
	err = c.session.DeletePacket(incoming, packetID)
	if err != nil {
		return c.die(SessionError, err, true)
	}

	return nil
}

// handle an incoming DisconnectPacket
func (c *Client) processDisconnect() error {
	// clear will
	err := c.session.ClearWill()
	if err != nil {
		return c.die(SessionError, err, true)
	}

	// close client
	c.Close(true)

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
			sub, err := c.session.LookupSubscription(publish.Message.Topic)
			if err != nil {
				return c.die(SessionError, err, true)
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
					return c.die(SessionError, err, true)
				}
			}

			// send packet
			err = c.send(publish, true)
			if err != nil {
				return c.die(TransportError, err, false)
			}

			c.log(MessageForwarded, c, nil, msg, nil)
		}
	}
}

/* helpers */

func (c *Client) handleMessage(msg *packet.Message) error {
	// check retain flag
	if msg.Retain {
		if len(msg.Payload) > 0 {
			err := c.engine.Backend.StoreRetained(c, msg)
			if err != nil {
				return err
			}
		} else {
			err := c.engine.Backend.ClearRetained(c, msg.Topic)
			if err != nil {
				return err
			}
		}
	}

	// reset an existing retain flag
	msg.Retain = false

	// publish message to others
	err := c.engine.Backend.Publish(c, msg)
	if err != nil {
		return err
	}

	c.log(MessagePublished, c, nil, msg, nil)

	return nil
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup(event LogEvent, err error, close bool) (LogEvent, error) {
	// check session
	if c.session != nil && c.state.get() == clientConnected {
		// get will
		will, willErr := c.session.LookupWill()
		if willErr != nil && err == nil {
			event = SessionError
			err = willErr
		}

		// publish will message
		if will != nil {
			willErr = c.handleMessage(will)
			if willErr != nil && err == nil {
				event = BackendError
				err = willErr
			}
		}
	}

	// remove client from the queue
	if c.state.get() > clientConnecting {
		termErr := c.engine.Backend.Terminate(c)
		if termErr != nil && err == nil {
			event = BackendError
			err = termErr
		}

		// reset the session if clean is requested
		if c.session != nil && c.cleanSession {
			resetErr := c.session.Reset()
			if resetErr != nil && err == nil {
				event = SessionError
				err = resetErr
			}
		}
	}

	// ensure that the connection gets closed
	if close {
		closeErr := c.conn.Close()
		if closeErr != nil && err == nil {
			event = TransportError
			err = closeErr
		}
	}

	c.log(LostConnection, c, nil, nil, nil)

	// remove client from the brokers list if added
	if c.state.get() > clientConnecting {
		c.engine.remove(c)
	}

	return event, err
}

// used for closing and cleaning up from internal goroutines
func (c *Client) die(event LogEvent, err error, close bool) error {
	c.finish.Do(func() {
		event, err = c.cleanup(event, err, close)

		// report error
		if err != nil {
			c.log(event, c, nil, nil, err)
		}
	})

	return err
}

func (c *Client) send(pkt packet.Packet, buffered bool) error {
	var err error

	// send packet
	if buffered {
		err = c.conn.BufferedSend(pkt)
	} else {
		err = c.conn.Send(pkt)
	}

	// check error
	if err != nil {
		return err
	}

	c.log(PacketSent, c, pkt, nil, nil)

	return nil
}

// log a message
func (c *Client) log(event LogEvent, client *Client, pkt packet.Packet, msg *packet.Message, err error) {
	if c.engine.Logger != nil {
		c.engine.Logger(event, client, pkt, msg, err)
	}
}
