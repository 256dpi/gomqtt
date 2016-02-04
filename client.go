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

// Message encapsulates a Topic and a Payload and is returned to the Callback
// when received from a broker.
type Message struct {
	Topic   string
	Payload []byte
}

// ErrAlreadyConnecting is returned by Connect if there was already a connection
// attempt.
var ErrAlreadyConnecting = errors.New("already connecting")

// ErrNotConnected is returned by Publish, Subscribe and Unsubscribe if the
// client is not currently connected.
var ErrNotConnected = errors.New("not connected")

// ErrMissingClientID is returned by Connect if no ClientID has been provided in
// the options while requesting an unclean session.
var ErrMissingClientID = errors.New("missing client id")

// ErrConnectionDenied is returned in the Callback if the connection has been
// reject by the broker.
var ErrConnectionDenied = errors.New("connection denied")

// ErrMissingPong is returned in the Callback if the broker did not respond to
// a PingreqPacket.
var ErrMissingPong = errors.New("missing pong")

// ErrUnexpectedClose is returned in the Callback if the broker closed the
// connection without receiving a DisconnectPacket from the client.
var ErrUnexpectedClose = errors.New("unexpected close")

// Callback is a function called by the client upon received messages or internal
// errors.
type Callback func(*Message, error)

// Logger is a function called by the client to log activity.
type Logger func(string)

// Client connects to a broker and handles the transmission of packets between them.
type Client struct {
	conn transport.Conn

	IncomingStore Store
	OutgoingStore Store
	Callback      Callback
	Logger        Logger

	state       *state
	counter     *counter
	tracker     *tracker
	futureStore *futureStore

	tomb  tomb.Tomb
	mutex sync.Mutex
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{
		IncomingStore: NewMemoryStore(),
		OutgoingStore: NewMemoryStore(),
		state:         newState(),
		counter:       newCounter(),
		futureStore:   newFutureStore(),
	}
}

/* exported interface */

// Connect opens the connection to the broker and sends a ConnectPacket. It will
// return a ConnectFuture that gets completed once a ConnackPacket has been
// received. If the ConnectPacket couldn't be transmitted it will return an error.
// It will return ErrAlreadyConnecting if Connect has been called before.
func (c *Client) Connect(urlString string, opts *Options) (*ConnectFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if already connecting
	if c.state.get() >= stateConnecting {
		return nil, ErrAlreadyConnecting
	}

	// parse url
	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	// save opts
	if opts == nil {
		opts = NewOptions("gomqtt/client")
	}

	// check client id
	if !opts.CleanSession && opts.ClientID == "" {
		return nil, ErrMissingClientID
	}

	// parse keep alive
	keepAlive, err := time.ParseDuration(opts.KeepAlive)
	if err != nil {
		return nil, err
	}

	// allocate and initialize tracker
	c.tracker = newTracker(keepAlive)

	// dial broker
	c.conn, err = transport.Dial(urlString)
	if err != nil {
		return nil, err
	}

	// set to connecting as from this point the client cannot be reused
	c.state.set(stateConnecting)

	// from now on the connection has been used and we have to close the
	// connection and cleanup on any subsequent error

	// open incoming store
	err = c.IncomingStore.Open(opts.CleanSession)
	if err != nil {
		return nil, c.cleanup(err, true)
	}

	// open outgoing store
	err = c.OutgoingStore.Open(opts.CleanSession)
	if err != nil {
		return nil, c.cleanup(err, true)
	}

	// allocate packet
	connect := packet.NewConnectPacket()
	connect.ClientID = []byte(opts.ClientID)
	connect.KeepAlive = uint16(keepAlive.Seconds())
	connect.CleanSession = opts.CleanSession

	// check for credentials
	if urlParts.User != nil {
		connect.Username = []byte(urlParts.User.Username())
		p, _ := urlParts.User.Password()
		connect.Password = []byte(p)
	}

	// set will
	connect.WillTopic = []byte(opts.WillTopic)
	connect.WillPayload = opts.WillPayload
	connect.WillQOS = opts.WillQos
	connect.WillRetain = opts.WillRetained

	// create new ConnackFuture
	future := &ConnectFuture{}
	future.initialize()

	// store future with id 0
	c.futureStore.put(c.counter.next(), future)

	// send connect packet
	err = c.send(connect)
	if err != nil {
		return nil, c.cleanup(err, false)
	}

	// start process routine
	c.tomb.Go(c.processor)

	// start keep alive if greater than zero
	if keepAlive > 0 {
		c.tomb.Go(c.pinger)
	}

	return future, nil
}

// Publish will send a PublishPacket containing the passed parameters.
func (c *Client) Publish(topic string, payload []byte, qos byte, retain bool) (*PublishFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != stateConnected {
		return nil, ErrNotConnected
	}

	// allocate packet
	publish := packet.NewPublishPacket()
	publish.Topic = []byte(topic)
	publish.Payload = payload
	publish.QOS = qos
	publish.Retain = retain
	publish.Dup = false

	// set packet id
	if qos > 0 {
		publish.PacketID = c.counter.next()
	}

	// store packet
	if qos >= 1 {
		err := c.OutgoingStore.Put(publish)
		if err != nil {
			return nil, c.cleanup(err, true)
		}
	}

	// create future
	future := &PublishFuture{}
	future.initialize()

	// store future
	c.futureStore.put(publish.PacketID, future)

	// send packet
	err := c.send(publish)
	if err != nil {
		return nil, c.cleanup(err, false)
	}

	// complete and remove qos 1 future
	if qos == 0 {
		future.complete()
		c.futureStore.del(publish.PacketID)
	}

	return future, nil
}

// Subscribe will send a SubscribePacket containing one topic to subscribe.
func (c *Client) Subscribe(topic string, qos byte) (*SubscribeFuture, error) {
	return c.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

// SubscribeMultiple will send a SubscribePacket containing multiple topics to
// subscribe.
func (c *Client) SubscribeMultiple(filters map[string]byte) (*SubscribeFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != stateConnected {
		return nil, ErrNotConnected
	}

	// allocate packet
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = make([]packet.Subscription, 0, len(filters))
	subscribe.PacketID = c.counter.next()

	// append filters
	for topic, qos := range filters {
		subscribe.Subscriptions = append(subscribe.Subscriptions, packet.Subscription{
			Topic: []byte(topic),
			QOS:   qos,
		})
	}

	// store packet
	err := c.OutgoingStore.Put(subscribe)
	if err != nil {
		return nil, c.cleanup(err, true)
	}

	// create future
	future := &SubscribeFuture{}
	future.initialize()

	// store future
	c.futureStore.put(subscribe.PacketID, future)

	// send packet
	err = c.send(subscribe)
	if err != nil {
		return nil, c.cleanup(err, false)
	}

	return future, nil
}

// Unsubscribe will send a UnsubscribePacket containing one topic to unsubscribe.
func (c *Client) Unsubscribe(topic string) (*UnsubscribeFuture, error) {
	return c.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a UnsubscribePacket containing multiple
// topics to unsubscribe.
func (c *Client) UnsubscribeMultiple(topics []string) (*UnsubscribeFuture, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != stateConnected {
		return nil, ErrNotConnected
	}

	// allocate packet
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = make([][]byte, 0, len(topics))
	unsubscribe.PacketID = c.counter.next()

	// append topics
	for _, t := range topics {
		unsubscribe.Topics = append(unsubscribe.Topics, []byte(t))
	}

	// store packet
	err := c.OutgoingStore.Put(unsubscribe)
	if err != nil {
		return nil, c.cleanup(err, true)
	}

	// create future
	future := &UnsubscribeFuture{}
	future.initialize()

	// store future
	c.futureStore.put(unsubscribe.PacketID, future)

	// send packet
	err = c.send(unsubscribe)
	if err != nil {
		return nil, c.cleanup(err, false)
	}

	return future, nil
}

// Disconnect will send a DisconnectPacket and close the connection.
func (c *Client) Disconnect(timeout ...time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if connected
	if c.state.get() != stateConnected {
		return ErrNotConnected
	}

	// allocate packet
	m := packet.NewDisconnectPacket()

	// set state
	c.state.set(stateDisconnecting)

	// finish current packets
	if len(timeout) > 0 {
		c.futureStore.await(timeout[0])
	}

	// send disconnect packet
	err := c.send(m)

	// shutdown goroutines
	c.tomb.Kill(nil)

	// wait for all goroutines to exit
	c.tomb.Wait()

	// do cleanup
	return c.cleanup(err, false)
}

/* processor goroutine */

// processes incoming packets
func (c *Client) processor() error {
	for {
		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		default:
			// get next packet from connection
			pkt, err := c.conn.Receive()
			if err != nil {
				transportErr, ok := err.(transport.Error)

				if ok && transportErr.Code() == transport.ExpectedClose {
					if c.state.get() != stateDisconnecting {
						return c.die(ErrUnexpectedClose, false)
					}

					// connection has been closed because of a clean Disonnect
					return nil
				}

				// die on any other error
				return c.die(err, false)
			}

			// TODO: Lock processor as well?

			c.log("Received: %s", pkt.String())

			// call handlers for packet types
			switch pkt.Type() {
			case packet.CONNACK:
				err = c.processConnack(pkt.(*packet.ConnackPacket))
			case packet.SUBACK:
				err = c.processSuback(pkt.(*packet.SubackPacket))
			case packet.UNSUBACK:
				err = c.processUnsuback(pkt.(*packet.UnsubackPacket))
			case packet.PINGRESP:
				c.tracker.pong()
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
			default:
				// ignore unsupported packet types
			}

			// return eventual error
			if err != nil {
				return err // error has already been cleaned
			}
		}
	}
}

// handle the incoming ConnackPacket
func (c *Client) processConnack(connack *packet.ConnackPacket) error {
	// check state
	if c.state.get() != stateConnecting {
		return nil // ignore wrongly sent ConnackPacket
	}

	// get future
	connectFuture, ok := c.futureStore.get(0).(*ConnectFuture)
	if !ok {
		// must be available otherwise the broker messed completely up...
		return nil
	}

	// fill future
	connectFuture.SessionPresent = connack.SessionPresent
	connectFuture.ReturnCode = connack.ReturnCode

	// remove future
	c.futureStore.del(0)

	// return connection denied error and close connection if not accepted
	if connack.ReturnCode != packet.ConnectionAccepted {
		err := c.die(ErrConnectionDenied, true)
		connectFuture.complete()
		return err
	}

	// set state to connected
	c.state.set(stateConnected)

	// complete future
	connectFuture.complete()

	// retrieve stored packets
	packets, err := c.OutgoingStore.All()
	if err != nil {
		return c.die(err, true)
	}

	// resend stored packets
	for _, pkt := range packets {
		err = c.send(pkt)
		if err != nil {
			return c.die(err, false)
		}
	}

	return nil
}

// handle an incoming SubackPacket
func (c *Client) processSuback(suback *packet.SubackPacket) error {
	// remove packet from store
	c.OutgoingStore.Del(suback.PacketID)

	// find future
	future := c.futureStore.get(suback.PacketID)

	// check future
	subscribeFuture, ok := future.(*SubscribeFuture)
	if !ok {
		return nil // ignore a wrongly sent SubackPacket
	}

	// complete future
	subscribeFuture.ReturnCodes = suback.ReturnCodes
	subscribeFuture.complete()

	// remove future from store
	c.futureStore.del(suback.PacketID)

	return nil
}

// handle an incoming UnsubackPacket
func (c *Client) processUnsuback(unsuback *packet.UnsubackPacket) error {
	// remove packet from store
	c.OutgoingStore.Del(unsuback.PacketID)

	// find future
	future := c.futureStore.get(unsuback.PacketID)

	// check future
	unsubscribeFuture, ok := future.(*UnsubscribeFuture)
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
		err := c.IncomingStore.Put(publish)
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
		// call callback
		c.forward(publish)
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) processPubackAndPubcomp(packetID uint16) error {
	// remove packet from store
	c.OutgoingStore.Del(packetID)

	// find future
	future := c.futureStore.get(packetID)

	// check future
	publishFuture, ok := future.(*PublishFuture)
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
	// allocate packet
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	// overwrite stored PublishPacket with new PubRelPacket
	err := c.OutgoingStore.Put(pubrel)
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
	pkt, err := c.IncomingStore.Get(packetID)
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
	err = c.IncomingStore.Del(packetID)
	if err != nil {
		return c.die(err, true)
	}

	// call callback
	c.forward(publish)

	return nil
}

/* pinger goroutine */

// manages the sending of ping packets to keep the connection alive
func (c *Client) pinger() error {
	for {
		window := c.tracker.window()

		// TODO: Lock pinger as well?

		if window < 0 {
			if c.tracker.pending() {
				return c.die(ErrMissingPong, true)
			}

			err := c.send(packet.NewPingreqPacket())
			if err != nil {
				return c.die(err, false)
			}

			c.tracker.ping()
		} else {
			c.log(fmt.Sprintf("Delay KeepAlive by %s", window.String()))
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

// sends message and updates lastSend
func (c *Client) send(pkt packet.Packet) error {
	c.tracker.reset()

	// send packet
	err := c.conn.Send(pkt)
	if err != nil {
		return err
	}

	c.log("Sent: %s", pkt.String())

	return nil
}

// calls the callback with a new message
func (c *Client) forward(packet *packet.PublishPacket) {
	msg := newMessage(string(packet.Topic), packet.Payload)

	if c.Callback != nil {
		c.Callback(msg, nil)
	}
}

// log a message
func (c *Client) log(format string, a ...interface{}) {
	if c.Logger != nil {
		c.Logger(fmt.Sprintf(format, a...))
	}
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup(err error, close bool) error {
	// set state
	c.state.set(stateDisconnected)

	// ensure that the connection gets closed
	if close {
		_err := c.conn.Close()
		if err == nil {
			err = _err
		}
	}

	// close incoming store
	_err := c.IncomingStore.Close()
	if err == nil {
		err = _err
	}

	// close outgoing store
	_err = c.OutgoingStore.Close()
	if err == nil {
		err = _err
	}

	return err
}

// used for closing and cleaning up from inside internal goroutines
func (c *Client) die(err error, close bool) error {
	err = c.cleanup(err, close)

	if c.Callback != nil {
		c.Callback(nil, err)
	}

	return err
}
