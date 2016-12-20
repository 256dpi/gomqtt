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

package client

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"gopkg.in/tomb.v2"
)

// ErrClientAlreadyConnecting is returned by Connect if there has been already a
// connection attempt.
var ErrClientAlreadyConnecting = errors.New("client already connecting")

// ErrClientNotConnected is returned by Publish, Subscribe and Unsubscribe if the
// client is not currently connected.
var ErrClientNotConnected = errors.New("client not connected")

// ErrClientMissingID is returned by Connect if no ClientID has been provided in
// the config while requesting to resume a session.
var ErrClientMissingID = errors.New("client missing id")

// ErrClientConnectionDenied is returned in the Callback if the connection has
// been reject by the broker.
var ErrClientConnectionDenied = errors.New("client connection denied")

// ErrClientMissingPong is returned in the Callback if the broker did not respond
// in time to a PingreqPacket.
var ErrClientMissingPong = errors.New("client missing pong")

// ErrClientExpectedConnack is returned when the first received packet is not a
// ConnackPacket.
var ErrClientExpectedConnack = errors.New("client expected connack")

// ErrFailedSubscription is returned when a submitted subscription is marked as
// failed when Config.ValidateSubs must be set to true.
var ErrFailedSubscription = errors.New("failed subscription")

// A Callback is a function called by the client upon received messages or
// internal errors.
//
// Note: Execution of the client is resumed after the callback returns. This
// means that waiting on a future will actually deadlock the client.
type Callback func(msg *packet.Message, err error)

// A Logger is a function called by the client to log activity.
type Logger func(msg string)

const (
	clientInitialized byte = iota
	clientConnecting
	clientConnacked
	clientConnected
	clientDisconnecting
	clientDisconnected
)

// A Client connects to a broker and handles the transmission of packets. It will
// automatically send PingreqPackets to keep the connection alive. Outgoing
// publish related packets will be stored in session and resent when the
// connection gets closed abruptly. All methods return Futures that get completed
// when the packets get acknowledged by the broker. Once the connection is closed
// all waiting futures get canceled.
//
// Note: If clean session is set to false and there are packets in the session,
// messages might get completed after connecting without triggering any futures
// to complete.
type Client struct {
	config *Config
	conn   transport.Conn

	// The session used by the client to store unacknowledged packets.
	Session Session

	// The callback to be called by the client upon receiving a message or
	// encountering an error while processing incoming packets.
	Callback Callback

	// The logger that is used to log write low level information like packets
	// that have ben successfully sent and received and details about the
	// automatic keep alive handler.
	Logger Logger

	state   *state
	tracker *tracker
	clean   bool

	futureStore   *futureStore
	connectFuture *ConnectFuture

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// New returns a new client that by default uses a fresh MemorySession.
func New() *Client {
	return &Client{
		Session:     NewMemorySession(),
		state:       newState(clientInitialized),
		futureStore: newFutureStore(),
	}
}

// Connect opens the connection to the broker and sends a ConnectPacket. It will
// return a ConnectFuture that gets completed once a ConnackPacket has been
// received. If the ConnectPacket couldn't be transmitted it will return an error.
func (c *Client) Connect(config *Config) (*ConnectFuture, error) {
	if config == nil {
		panic("No config specified")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// save config
	c.config = config

	// check if already connecting
	if c.state.get() >= clientConnecting {
		return nil, ErrClientAlreadyConnecting
	}

	// parse url
	urlParts, err := url.ParseRequestURI(config.BrokerURL)
	if err != nil {
		return nil, err
	}

	// check client id
	if !config.CleanSession && config.ClientID == "" {
		return nil, ErrClientMissingID
	}

	// parse keep alive
	keepAlive, err := time.ParseDuration(config.KeepAlive)
	if err != nil {
		return nil, err
	}

	// allocate and initialize tracker
	c.tracker = newTracker(keepAlive)

	// dial broker (with custom dialer if present)
	if config.Dialer != nil {
		c.conn, err = config.Dialer.Dial(config.BrokerURL)
		if err != nil {
			return nil, err
		}
	} else {
		c.conn, err = transport.Dial(config.BrokerURL)
		if err != nil {
			return nil, err
		}
	}

	// set to connecting as from this point the client cannot be reused
	c.state.set(clientConnecting)

	// from now on the connection has been used and we have to close the
	// connection and cleanup on any subsequent error

	// save clean
	c.clean = config.CleanSession

	// reset store
	if c.clean {
		err = c.Session.Reset()
		if err != nil {
			return nil, c.cleanup(err, true, false)
		}
	}

	// allocate packet
	connect := packet.NewConnectPacket()
	connect.ClientID = config.ClientID
	connect.KeepAlive = uint16(keepAlive.Seconds())
	connect.CleanSession = config.CleanSession

	// check for credentials
	if urlParts.User != nil {
		connect.Username = urlParts.User.Username()
		connect.Password, _ = urlParts.User.Password()
	}

	// set will
	connect.Will = config.WillMessage

	// create new ConnackFuture
	c.connectFuture = &ConnectFuture{}
	c.connectFuture.initialize()

	// send connect packet
	err = c.send(connect, false)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	// start process routine
	c.tomb.Go(c.processor)

	// start keep alive if greater than zero
	if keepAlive > 0 {
		c.tomb.Go(c.pinger)
	}

	return c.connectFuture, nil
}

// Publish will send a PublishPacket containing the passed parameters. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (c *Client) Publish(topic string, payload []byte, qos uint8, retain bool) (*PublishFuture, error) {
	msg := &packet.Message{
		Topic:   topic,
		Payload: payload,
		QOS:     qos,
		Retain:  retain,
	}

	return c.PublishMessage(msg)
}

// PublishMessage will send a PublishPacket containing the passed message. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (c *Client) PublishMessage(msg *packet.Message) (*PublishFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != clientConnected {
		return nil, ErrClientNotConnected
	}

	// allocate packet
	publish := packet.NewPublishPacket()
	publish.Message = *msg

	// set packet id
	if msg.QOS > 0 {
		publish.PacketID = c.Session.PacketID()
	}

	// create future
	future := &PublishFuture{}
	future.initialize()

	// store future
	c.futureStore.put(publish.PacketID, future)

	// store packet if at least qos 1
	if msg.QOS > 0 {
		err := c.Session.SavePacket(outgoing, publish)
		if err != nil {
			return nil, c.cleanup(err, true, false)
		}
	}

	// send packet
	err := c.send(publish, true)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	// complete and remove qos 0 future
	if msg.QOS == 0 {
		future.complete()
		c.futureStore.del(publish.PacketID)
	}

	return future, nil
}

// Subscribe will send a SubscribePacket containing one topic to subscribe. It
// will return a SubscribeFuture that gets completed once a SubackPacket has
// been received.
func (c *Client) Subscribe(topic string, qos uint8) (*SubscribeFuture, error) {
	return c.SubscribeMultiple([]packet.Subscription{
		{Topic: topic, QOS: qos},
	})
}

// SubscribeMultiple will send a SubscribePacket containing multiple topics to
// subscribe. It will return a SubscribeFuture that gets completed once a
// SubackPacket has been received.
func (c *Client) SubscribeMultiple(subscriptions []packet.Subscription) (*SubscribeFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != clientConnected {
		return nil, ErrClientNotConnected
	}

	// allocate packet
	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = c.Session.PacketID()
	subscribe.Subscriptions = subscriptions

	// create future
	future := &SubscribeFuture{}
	future.initialize()

	// store future
	c.futureStore.put(subscribe.PacketID, future)

	// send packet
	err := c.send(subscribe, true)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	return future, nil
}

// Unsubscribe will send a UnsubscribePacket containing one topic to unsubscribe.
// It will return a UnsubscribeFuture that gets completed once a UnsubackPacket
// has been received.
func (c *Client) Unsubscribe(topic string) (*UnsubscribeFuture, error) {
	return c.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a UnsubscribePacket containing multiple
// topics to unsubscribe. It will return a UnsubscribeFuture that gets completed
// once a UnsubackPacket has been received.
func (c *Client) UnsubscribeMultiple(topics []string) (*UnsubscribeFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != clientConnected {
		return nil, ErrClientNotConnected
	}

	// allocate packet
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = topics
	unsubscribe.PacketID = c.Session.PacketID()

	// create future
	future := &UnsubscribeFuture{}
	future.initialize()

	// store future
	c.futureStore.put(unsubscribe.PacketID, future)

	// send packet
	err := c.send(unsubscribe, true)
	if err != nil {
		return nil, c.cleanup(err, false, false)
	}

	return future, nil
}

// Disconnect will send a DisconnectPacket and close the connection.
//
// If a timeout is specified, the client will wait the specified amount of time
// for all queued futures to complete or cancel. If no timeout is specified it
// will not wait at all.
func (c *Client) Disconnect(timeout ...time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != clientConnected {
		return ErrClientNotConnected
	}

	// finish current packets
	if len(timeout) > 0 {
		c.futureStore.await(timeout[0])
	}

	// set state
	c.state.set(clientDisconnecting)

	// send disconnect packet
	err := c.send(packet.NewDisconnectPacket(), false)

	return c.end(err, true)
}

// Close closes the client immediately without sending a DisconnectPacket and
// waiting for outgoing transmissions to finish.
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() < clientConnecting {
		return ErrClientNotConnected
	}

	return c.end(nil, false)
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	first := true

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			// if we are disconnecting we can ignore the error
			if c.state.get() >= clientDisconnecting {
				return nil
			}

			// die on any other error
			return c.die(err, false)
		}

		// log received message
		if c.Logger != nil {
			c.Logger(fmt.Sprintf("Received: %s", pkt.String()))
		}

		if first {
			// get connack
			connack, ok := pkt.(*packet.ConnackPacket)
			if !ok {
				return c.die(ErrClientExpectedConnack, true)
			}

			// process connack
			err = c.processConnack(connack)
			first = false

			// move on
			continue
		}

		// call handlers for packet types and ignore other packets
		switch _pkt := pkt.(type) {
		case *packet.SubackPacket:
			err = c.processSuback(_pkt)
		case *packet.UnsubackPacket:
			err = c.processUnsuback(_pkt)
		case *packet.PingrespPacket:
			c.tracker.pong()
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
		}

		// return eventual error
		if err != nil {
			return err // error has already been cleaned
		}
	}
}

// handle the incoming ConnackPacket
func (c *Client) processConnack(connack *packet.ConnackPacket) error {
	// check state
	if c.state.get() != clientConnecting {
		return nil // ignore wrongly sent ConnackPacket
	}

	// set state
	c.state.set(clientConnacked)

	// fill future
	c.connectFuture.SessionPresent = connack.SessionPresent
	c.connectFuture.ReturnCode = connack.ReturnCode

	// return connection denied error and close connection if not accepted
	if connack.ReturnCode != packet.ConnectionAccepted {
		err := c.die(ErrClientConnectionDenied, true)
		c.connectFuture.complete()
		return err
	}

	// set state to connected
	c.state.set(clientConnected)

	// complete future
	c.connectFuture.complete()

	// retrieve stored packets
	packets, err := c.Session.AllPackets(outgoing)
	if err != nil {
		return c.die(err, true)
	}

	// resend stored packets
	for _, pkt := range packets {
		// check for publish packets
		publish, ok := pkt.(*packet.PublishPacket)
		if ok {
			// set the dup flag on a publish packet
			publish.Dup = true
		}

		// resend packet
		err = c.send(pkt, true)
		if err != nil {
			return c.die(err, false)
		}
	}

	return nil
}

// handle an incoming SubackPacket
func (c *Client) processSuback(suback *packet.SubackPacket) error {
	// remove packet from store
	err := c.Session.DeletePacket(outgoing, suback.PacketID)
	if err != nil {
		return err
	}

	// get future
	subscribeFuture, ok := c.futureStore.get(suback.PacketID).(*SubscribeFuture)
	if !ok {
		return nil // ignore a wrongly sent SubackPacket
	}

	// remove future from store
	c.futureStore.del(suback.PacketID)

	// validate subscriptions if requested
	if c.config.ValidateSubs {
		for _, code := range suback.ReturnCodes {
			if code == packet.QOSFailure {
				subscribeFuture.cancel()
				return ErrFailedSubscription
			}
		}
	}

	// complete future
	subscribeFuture.ReturnCodes = suback.ReturnCodes
	subscribeFuture.complete()

	return nil
}

// handle an incoming UnsubackPacket
func (c *Client) processUnsuback(unsuback *packet.UnsubackPacket) error {
	// remove packet from store
	err := c.Session.DeletePacket(outgoing, unsuback.PacketID)
	if err != nil {
		return err
	}

	// get future
	unsubscribeFuture, ok := c.futureStore.get(unsuback.PacketID).(*UnsubscribeFuture)
	if !ok {
		return nil // ignore a wrongly sent UnsubackPacket
	}

	// complete future
	unsubscribeFuture.complete()

	// remove future from store
	c.futureStore.del(unsuback.PacketID)

	return nil
}

// handle an incoming PublishPacket
func (c *Client) processPublish(publish *packet.PublishPacket) error {
	if publish.Message.QOS == 1 {
		// prepare puback packet
		puback := packet.NewPubackPacket()
		puback.PacketID = publish.PacketID

		// acknowledge qos 1 publish
		err := c.send(puback, true)
		if err != nil {
			return c.die(err, false)
		}
	}

	if publish.Message.QOS == 2 {
		// store packet
		err := c.Session.SavePacket(incoming, publish)
		if err != nil {
			return c.die(err, true)
		}

		// prepare pubrec packet
		pubrec := packet.NewPubrecPacket()
		pubrec.PacketID = publish.PacketID

		// acknowledge qos 2 publish
		err = c.send(pubrec, true)
		if err != nil {
			return c.die(err, false)
		}
	}

	if publish.Message.QOS <= 1 {
		// call callback
		if c.Callback != nil {
			c.Callback(&publish.Message, nil)
		}
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) processPubackAndPubcomp(packetID uint16) error {
	// remove packet from store
	err := c.Session.DeletePacket(outgoing, packetID)
	if err != nil {
		return err
	}

	// get future
	publishFuture, ok := c.futureStore.get(packetID).(*PublishFuture)
	if !ok {
		return nil // ignore a wrongly sent PubackPacket or PubcompPacket
	}

	// complete future
	publishFuture.complete()

	// remove future from store
	c.futureStore.del(packetID)

	return nil
}

// handle an incoming PubrecPacket
func (c *Client) processPubrec(packetID uint16) error {
	// prepare pubrel packet
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	// overwrite stored PublishPacket with PubrelPacket
	err := c.Session.SavePacket(outgoing, pubrel)
	if err != nil {
		return c.die(err, true)
	}

	// send packet
	err = c.send(pubrel, true)
	if err != nil {
		return c.die(err, false)
	}

	return nil
}

// handle an incoming PubrelPacket
func (c *Client) processPubrel(packetID uint16) error {
	// get packet from store
	pkt, err := c.Session.LookupPacket(incoming, packetID)
	if err != nil {
		return c.die(err, true)
	}

	// get packet from store
	publish, ok := pkt.(*packet.PublishPacket)
	if !ok {
		return nil // ignore a wrongly sent PubrelPacket
	}

	// prepare pubcomp packet
	pubcomp := packet.NewPubcompPacket()
	pubcomp.PacketID = publish.PacketID

	// acknowledge PublishPacket
	err = c.send(pubcomp, true)
	if err != nil {
		return c.die(err, false)
	}

	// remove packet from store
	err = c.Session.DeletePacket(incoming, packetID)
	if err != nil {
		return c.die(err, true)
	}

	// call callback
	if c.Callback != nil {
		c.Callback(&publish.Message, nil)
	}

	return nil
}

/* pinger goroutine */

// manages the sending of ping packets to keep the connection alive
func (c *Client) pinger() error {
	for {
		// get current window
		window := c.tracker.window()

		// check if ping is due
		if window < 0 {
			// check if a pong has already been sent
			if c.tracker.pending() {
				return c.die(ErrClientMissingPong, true)
			}

			// send pingreq packet
			err := c.send(packet.NewPingreqPacket(), true)
			if err != nil {
				return c.die(err, false)
			}

			// save ping attempt
			c.tracker.ping()
		} else {
			// log keep alive delay
			if c.Logger != nil {
				c.Logger(fmt.Sprintf("Delay KeepAlive by %s", window.String()))
			}
		}

		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		case <-time.After(window):
			continue
		}
	}
}

/* helpers */

// sends packet and updates lastSend
func (c *Client) send(pkt packet.Packet, buffered bool) error {
	// reset keep alive tracker
	c.tracker.reset()

	// prepare error
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

	// log sent packet
	if c.Logger != nil {
		c.Logger(fmt.Sprintf("Sent: %s", pkt.String()))
	}

	return nil
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup(err error, doClose bool, possiblyClosed bool) error {
	// cancel connect future if appropriate
	if c.state.get() < clientConnacked && c.connectFuture != nil {
		c.connectFuture.cancel()
	}

	// set state
	c.state.set(clientDisconnected)

	// ensure that the connection gets closed
	if doClose {
		_err := c.conn.Close()
		if err == nil && !possiblyClosed {
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

	// cancel all futures
	c.futureStore.clear()

	return err
}

// used for closing and cleaning up from inside internal goroutines
func (c *Client) die(err error, close bool) error {
	c.finish.Do(func() {
		err = c.cleanup(err, close, false)

		if c.Callback != nil {
			c.Callback(nil, err)
		}
	})

	return err
}

// called by Disconnect and Close
func (c *Client) end(err error, possiblyClosed bool) error {
	// close connection
	err = c.cleanup(err, true, true)

	// shutdown goroutines
	c.tomb.Kill(nil)

	// wait for all goroutines to exit
	// goroutines will send eventual errors through the callback
	c.tomb.Wait()

	// do cleanup
	return err
}
