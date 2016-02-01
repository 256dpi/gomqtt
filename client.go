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

var ErrAlreadyConnecting = errors.New("already connecting")
var ErrMissingClientID = errors.New("missing client id")
var ErrAlreadyDisconnecting = errors.New("already disconnecting")
var ErrInvalidPacketType = errors.New("invalid packet type")

type Callback func(string, []byte, error)
type Logger func(string)

type Client struct {
	sync.Mutex

	IncomingStore Store
	OutgoingStore Store
	Callback      Callback
	Logger        Logger

	conn        transport.Conn
	counter     *counter
	futureStore *futureStore

	connecting    bool
	disconnecting bool
	connectFuture *ConnectFuture
	cleanSession  bool

	keepAlive       time.Duration
	lastSend        time.Time
	lastSendMutex   sync.Mutex
	pingrespPending bool

	tomb tomb.Tomb
	boot sync.WaitGroup
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{
		IncomingStore: NewMemoryStore(),
		OutgoingStore: NewMemoryStore(),
		futureStore:   newFutureStore(),
		counter:       newCounter(),
	}
}

// Connect opens the connection to the broker and sends a ConnectPacket. It will
// return a ConnectFuture that gets completed once a ConnackPacket has been
// received. If the ConnectPacket couldn't be transmitted it will return an error.
// It will return ErrAlreadyConnecting if Connect has been called before.
func (c *Client) Connect(urlString string, opts *Options) (*ConnectFuture, error) {
	c.Lock()
	defer c.Unlock()

	// check if already connecting
	if c.connecting {
		return nil, ErrAlreadyConnecting
	}

	// parse url
	urlParts, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	// save opts
	if opts == nil {
		opts = NewOptions("gomqtt/client")
	}

	// save clean session
	c.cleanSession = opts.CleanSession

	// check client id
	if !opts.CleanSession && opts.ClientID == "" {
		return nil, ErrMissingClientID
	}

	// parse keep alive
	c.keepAlive, err = time.ParseDuration(opts.KeepAlive)
	if err != nil {
		return nil, err
	}

	// dial broker
	c.conn, err = transport.Dial(urlString)
	if err != nil {
		return nil, err
	}

	// from now on we have to cleanup on a subsequent error

	// open incoming store
	err = c.IncomingStore.Open()
	if err != nil {
		return nil, c.cleanup(err)
	}

	// open outgoing store
	err = c.OutgoingStore.Open()
	if err != nil {
		return nil, c.cleanup(err)
	}

	// reset stores when clean session is set
	if opts.CleanSession {
		err = c.resetStores()
		if err != nil {
			return nil, c.cleanup(err)
		}
	}

	// allocate packet
	connect := packet.NewConnectPacket()
	connect.ClientID = []byte(opts.ClientID)
	connect.KeepAlive = uint16(c.keepAlive.Seconds())
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

	// send connect packet
	err = c.send(connect)
	if err != nil {
		return nil, c.cleanup(err)
	}

	// create new ConnackFuture
	future := &ConnectFuture{}
	future.initialize()

	// store future
	c.connectFuture = future

	// start process routine
	c.boot.Add(1)
	c.tomb.Go(c.process)

	// start keep alive if greater than zero
	if c.keepAlive > 0 {
		c.boot.Add(1)
		c.tomb.Go(c.ping)
	}

	// set state
	c.connecting = true

	// wait for all goroutines to start
	c.boot.Wait()

	return future, nil
}

// Publish will send a PublishPacket containing the passed parameters.
func (c *Client) Publish(topic string, payload []byte, qos byte, retain bool) (*PublishFuture, error) {
	c.Lock()
	defer c.Unlock()

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
			return nil, c.cleanup(err)
		}
	}

	// send packet
	err := c.send(publish)
	if err != nil {
		return nil, c.cleanup(err)
	}

	// create future
	future := &PublishFuture{}
	future.initialize()

	if qos == 0 {
		// instantly complete future
		future.complete()
	} else {
		// store future
		c.futureStore.put(publish.PacketID, future)
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
	c.Lock()
	defer c.Unlock()

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
		return nil, c.cleanup(err)
	}

	// send packet
	err = c.send(subscribe)
	if err != nil {
		return nil, c.cleanup(err)
	}

	// create future
	future := &SubscribeFuture{}
	future.initialize()

	// store future
	c.futureStore.put(subscribe.PacketID, future)

	return future, nil
}

// Unsubscribe will send a UnsubscribePacket containing one topic to unsubscribe.
func (c *Client) Unsubscribe(topic string) (*UnsubscribeFuture, error) {
	return c.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a UnsubscribePacket containing multiple
// topics to unsubscribe.
func (c *Client) UnsubscribeMultiple(topics []string) (*UnsubscribeFuture, error) {
	c.Lock()
	defer c.Unlock()

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
		return nil, c.cleanup(err)
	}

	// send packet
	err = c.send(unsubscribe)
	if err != nil {
		return nil, c.cleanup(err)
	}

	// create future
	future := &UnsubscribeFuture{}
	future.initialize()

	// store future
	c.futureStore.put(unsubscribe.PacketID, future)

	return future, nil
}

// Disconnect will send a DisconnectPacket and close the connection.
func (c *Client) Disconnect() error {
	c.Lock()
	defer c.Unlock()

	// check if already disconnecting
	if c.disconnecting {
		return ErrAlreadyDisconnecting
	}

	// allocate packet
	m := packet.NewDisconnectPacket()

	// TODO: finish outstanding work (set timeout parameter)

	// send disconnect packet
	err := c.send(m)
	if err != nil {
		transportErr, ok := err.(transport.Error)

		if ok && transportErr.Code() == transport.ExpectedClose {
			err = nil // do not treat expected close as an error
		}
	}

	// do cleanup
	err = c.cleanup(err)

	// wait for all goroutines to exit
	c.tomb.Wait()

	return err
}

// process incoming packets
func (c *Client) process() error {
	c.boot.Done()

	for {
		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		default:
			// get next packet from connection
			pkt, err := c.conn.Receive()
			if err != nil {
				return c.cleanup(err)
			}

			switch pkt.Type() {
			case packet.CONNACK:
				c.handleConnack(pkt.(*packet.ConnackPacket))
			case packet.SUBACK:
				c.handleSuback(pkt.(*packet.SubackPacket))
			case packet.UNSUBACK:
				c.handleUnsuback(pkt.(*packet.UnsubackPacket))
			case packet.PINGRESP:
				c.handlePingresp()
			case packet.PUBLISH:
				err = c.handlePublish(pkt.(*packet.PublishPacket))
			case packet.PUBACK:
				c.handlePubackAndPubcomp(pkt.(*packet.PubackPacket).PacketID)
			case packet.PUBCOMP:
				c.handlePubackAndPubcomp(pkt.(*packet.PubcompPacket).PacketID)
			case packet.PUBREC:
				err = c.handlePubrec(pkt.(*packet.PubrecPacket).PacketID)
			case packet.PUBREL:
				err = c.handlePubrel(pkt.(*packet.PubrelPacket).PacketID)
			default:
				return c.cleanup(ErrInvalidPacketType)
			}

			// return and cleanup on error
			if err != nil {
				return c.cleanup(err)
			}
		}
	}
}

// handle the incoming ConnackPacket
func (c *Client) handleConnack(connack *packet.ConnackPacket) {
	if !c.connectFuture.Completed() {
		c.connectFuture.SessionPresent = connack.SessionPresent
		c.connectFuture.ReturnCode = connack.ReturnCode
		c.connectFuture.complete()
	} else {
		// ignore a wrongly sent ConnackPacket
	}
}

// handle an incoming SubackPacket
func (c *Client) handleSuback(suback *packet.SubackPacket) {
	// remove packet from store
	c.OutgoingStore.Del(suback.PacketID)

	// find future
	future := c.futureStore.get(suback.PacketID)

	// complete future
	if subscribeFuture, ok := future.(*SubscribeFuture); ok {
		subscribeFuture.ReturnCodes = suback.ReturnCodes
		subscribeFuture.complete()
		c.futureStore.del(suback.PacketID)
	} else {
		// ignore a wrongly sent SubackPacket
	}
}

// handle an incoming UnsubackPacket
func (c *Client) handleUnsuback(unsuback *packet.UnsubackPacket) {
	// remove packet from store
	c.OutgoingStore.Del(unsuback.PacketID)

	// find future
	future := c.futureStore.get(unsuback.PacketID)

	// complete future
	if unsubscribeFuture, ok := future.(*UnsubscribeFuture); ok {
		unsubscribeFuture.complete()
		c.futureStore.del(unsuback.PacketID)
	} else {
		// ignore a wrongly sent UnsubackPacket
	}
}

// handle an incoming Pingresp
func (c *Client) handlePingresp() {
	c.pingrespPending = false
	c.log("Received PingrespPacket")
}

// handle an incoming PublishPacket
func (c *Client) handlePublish(publish *packet.PublishPacket) error {
	if publish.QOS == 1 {
		puback := packet.NewPubackPacket()
		puback.PacketID = publish.PacketID

		// acknowledge qos 1 publish
		err := c.send(puback)
		if err != nil {
			return err
		}
	}

	if publish.QOS == 2 {
		// store packet
		err := c.IncomingStore.Put(publish)
		if err != nil {
			return err
		}

		pubrec := packet.NewPubrecPacket()
		pubrec.PacketID = publish.PacketID

		// signal qos 2 publish
		err = c.send(pubrec)
		if err != nil {
			return err
		}
	}

	if publish.QOS <= 1 {
		// call callback
		c.forward(publish)
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) handlePubackAndPubcomp(packetID uint16) {
	// remove packet from store
	c.OutgoingStore.Del(packetID)

	// find future
	future := c.futureStore.get(packetID)

	// complete future
	if publishFuture, ok := future.(*PublishFuture); ok {
		publishFuture.complete()
		c.futureStore.del(packetID)
	} else {
		// ignore a wrongly sent PubackPacket or PubcompPacket
	}
}

// handle an incoming PubrecPacket
func (c *Client) handlePubrec(packetID uint16) error {
	// allocate packet
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	// overwrite stored PublishPacket with new PubRelPacket
	c.OutgoingStore.Put(pubrel)

	// send packet
	return c.send(pubrel)
}

// handle an incoming PubrelPacket
func (c *Client) handlePubrel(packetID uint16) error {
	// get packet from store
	pkt, err := c.IncomingStore.Get(packetID)
	if err != nil {
		return err
	}

	// get packet from store
	publish, ok := pkt.(*packet.PublishPacket)
	if !ok {
		// ignore a wrongly sent PubrelPacket
	}

	pubcomp := packet.NewPubcompPacket()
	pubcomp.PacketID = publish.PacketID

	// acknowledge PublishPacket
	err = c.send(pubcomp)
	if err != nil {
		return err
	}

	// remove packet from store
	err = c.IncomingStore.Del(packetID)
	if err != nil {
		return err
	}

	// call callback
	c.forward(publish)

	return nil
}

// sends message and updates lastSend
func (c *Client) send(msg packet.Packet) error {
	c.lastSendMutex.Lock()
	c.lastSend = time.Now()
	c.lastSendMutex.Unlock()

	return c.conn.Send(msg)
}

// calls the callback with an error
func (c *Client) error(err error) {
	if c.Callback != nil {
		c.Callback("", nil, err)
	}
}

// calls the callback with a new message
func (c *Client) forward(packet *packet.PublishPacket) {
	if c.Callback != nil {
		c.Callback(string(packet.Topic), packet.Payload, nil)
	}
}

// log a message
func (c *Client) log(message string) {
	if c.Logger != nil {
		c.Logger(message)
	}
}

// manages the sending of ping packets to keep the connection alive
func (c *Client) ping() error {
	c.boot.Done()

	for {
		c.lastSendMutex.Lock()
		timeElapsed := time.Since(c.lastSend)
		c.lastSendMutex.Unlock()

		timeToWait := c.keepAlive

		if timeElapsed > c.keepAlive {
			err := c.send(packet.NewPingreqPacket())
			if err != nil {
				return c.cleanup(err)
			}

			c.log(fmt.Sprintf("Sent PingreqPacket"))
		} else {
			timeToWait = c.keepAlive - timeElapsed

			c.log(fmt.Sprintf("Delay KeepAlive by %s", timeToWait.String()))
		}

		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		case <-time.After(timeToWait):
			continue
		}
	}
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup(err error) error {
	// set state
	c.disconnecting = true

	// ensure that the connection gets closed
	_err := c.conn.Close()
	if err == nil {
		err = _err
	}

	// reset stores if client has connected with a clean session
	if c.cleanSession {
		_err = c.resetStores()
		if err == nil {
			err = _err
		}
	}

	// close incoming store
	_err = c.IncomingStore.Close()
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

// resets the incoming and outgoing store
func (c *Client) resetStores() error {
	// reset incoming store
	err := c.IncomingStore.Reset()
	if err != nil {
		return err
	}

	// reset outgoing store
	err = c.OutgoingStore.Reset()
	if err != nil {
		return err
	}

	return nil
}
