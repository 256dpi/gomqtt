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
var ErrInvalidPacketType = errors.New("invalid packet type")

type (
	MessageCallback func(string, []byte)
	ErrorCallback   func(error)
	LogCallback     func(string)
)

type Client struct {
	conn transport.Conn

	IncomingStore Store
	OutgoingStore Store

	futureStore *futureStore
	idGenerator *idGenerator

	messageCallback MessageCallback
	errorCallback   ErrorCallback
	logCallback     LogCallback

	keepAlive       time.Duration
	lastSend        time.Time
	lastSendMutex   sync.Mutex
	pingrespPending bool

	connectFuture *ConnectFuture
	connectMutex  sync.Mutex

	tomb tomb.Tomb
	boot sync.WaitGroup
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{
		IncomingStore: NewMemoryStore(),
		OutgoingStore: NewMemoryStore(),
		futureStore: newFutureStore(),
		idGenerator: newIDGenerator(),
	}
}

// OnMessage sets the callback for incoming messages.
func (c *Client) OnMessage(callback MessageCallback) {
	c.messageCallback = callback
}

// OnError sets the callback for failed connection attempts and parsing errors.
func (c *Client) OnError(callback ErrorCallback) {
	c.errorCallback = callback
}

// OnLog sets the callback for log messages.
func (c *Client) OnLog(callback LogCallback) {
	c.logCallback = callback
}

// Connect opens the connection to the broker and sends a ConnectPacket. It will
// return a ConnectFuture that gets completed once a ConnackPacket has been
// received. If the ConnectPacket couldn't be transmitted it will return an error.
// It will return ErrAlreadyConnecting if Connect has been called before.
func (c *Client) Connect(urlString string, opts *Options) (*ConnectFuture, error) {
	c.connectMutex.Lock()
	defer c.connectMutex.Unlock()

	// TODO: we might use another identifier for that?
	// check for existing ConnectFuture
	if c.connectFuture != nil {
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

	// prepare connect packet
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
	c.connectFuture = &ConnectFuture{}
	c.connectFuture.initialize()

	// start process routine
	c.boot.Add(1)
	c.tomb.Go(c.process)

	// start keep alive if greater than zero
	if c.keepAlive > 0 {
		c.boot.Add(1)
		c.tomb.Go(c.ping)
	}

	// wait for all goroutines to start
	c.boot.Wait()

	return c.connectFuture, nil
}

// Publish will send a PublishPacket containing the passed parameters.
func (c *Client) Publish(topic string, payload []byte, qos byte, retain bool) (*PublishFuture, error) {
	publish := packet.NewPublishPacket()
	publish.Topic = []byte(topic)
	publish.Payload = payload
	publish.QOS = qos
	publish.Retain = retain
	publish.Dup = false
	publish.PacketID = c.idGenerator.next()

	// TODO: store packet

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
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = make([]packet.Subscription, 0, len(filters))
	subscribe.PacketID = c.idGenerator.next()

	// append filters
	for topic, qos := range filters {
		subscribe.Subscriptions = append(subscribe.Subscriptions, packet.Subscription{
			Topic: []byte(topic),
			QOS:   qos,
		})
	}

	// TODO: store packet

	// send packet
	err := c.send(subscribe)
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
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = make([][]byte, 0, len(topics))
	unsubscribe.PacketID = c.idGenerator.next()

	// append topics
	for _, t := range topics {
		unsubscribe.Topics = append(unsubscribe.Topics, []byte(t))
	}

	// TODO: store packet

	// send packet
	err := c.send(unsubscribe)
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
				err := c.handlePublish(pkt.(*packet.PublishPacket))
				if err != nil {
					return c.cleanup(err)
				}
			case packet.PUBACK:
				c.handlePubackAndPubcomp(pkt.(*packet.PubackPacket).PacketID)
			case packet.PUBCOMP:
				c.handlePubackAndPubcomp(pkt.(*packet.PubcompPacket).PacketID)
			case packet.PUBREC:
				err := c.handlePubrec(pkt.(*packet.PubrecPacket).PacketID)
				if err != nil {
					return c.cleanup(err)
				}
			case packet.PUBREL:
				c.handlePubrel(pkt.(*packet.PubrelPacket).PacketID)
			default:
				return c.cleanup(ErrInvalidPacketType)
			}
		}
	}
}

// handle the incoming ConnackPacket
func (c *Client) handleConnack(connack *packet.ConnackPacket) {
	// TODO: return error if future is already completed?
	c.connectFuture.SessionPresent = connack.SessionPresent
	c.connectFuture.ReturnCode = connack.ReturnCode
	c.connectFuture.complete()
}

// handle an incoming SubackPacket
func (c *Client) handleSuback(suback *packet.SubackPacket) {
	future := c.futureStore.get(suback.PacketID)

	if subscribeFuture, ok := future.(*SubscribeFuture); ok {
		subscribeFuture.complete()
		c.futureStore.del(suback.PacketID)
	} else {
		// ignore a wrongly set SubackPacket
	}
}

// handle an incoming UnsubackPacket
func (c *Client) handleUnsuback(unsuback *packet.UnsubackPacket) {
	future := c.futureStore.get(unsuback.PacketID)

	if subscribeFuture, ok := future.(*UnsubscribeFuture); ok {
		subscribeFuture.complete()
		c.futureStore.del(unsuback.PacketID)
	} else {
		// ignore a wrongly set UnsubackPacket
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
			return c.cleanup(err)
		}
	}

	if publish.QOS == 2 {
		// TODO: store packet

		pubrec := packet.NewPubrecPacket()
		pubrec.PacketID = publish.PacketID

		// signal qos 2 publish
		err := c.send(pubrec)
		if err != nil {
			return c.cleanup(err)
		}
	}

	if publish.QOS <= 1 {
		// call callback
		if c.messageCallback != nil {
			c.message(publish)
		}
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) handlePubackAndPubcomp(packetID uint16) {
	future := c.futureStore.get(packetID)

	if publishFuture, ok := future.(*PublishFuture); ok {
		publishFuture.complete()
		c.futureStore.del(packetID)
	} else {
		// ignore a wrongly sent PubackPacket or PubcompPacket
	}

	// TODO: Remove Packet from store
}

// handle an incoming PubrecPacket
func (c *Client) handlePubrec(packetID uint16) error {
	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = packetID

	return c.send(pubrel)
}

// handle an incoming PubrelPacket
func (c *Client) handlePubrel(packetID uint16) {
	// Check store for matching PublishPacket
	// Call OnMessage callback
	// Save PubrelPacket to Store (why not remove it?)
	// Send Pubcomp packet
}

// sends message and updates lastSend
func (c *Client) send(msg packet.Packet) error {
	c.lastSendMutex.Lock()
	c.lastSend = time.Now()
	c.lastSendMutex.Unlock()

	return c.conn.Send(msg)
}

// calls the error callback
func (c *Client) error(err error) {
	if c.errorCallback != nil {
		c.errorCallback(err)
	}
}

// calls the log callback
func (c *Client) log(message string) {
	if c.logCallback != nil {
		c.logCallback(message)
	}
}

// calls the message callback
func (c *Client) message(packet *packet.PublishPacket) {
	if c.messageCallback != nil {
		c.messageCallback(string(packet.Topic), packet.Payload)
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
	// ensure that the connection gets closed
	_err := c.conn.Close()
	if err == nil {
		err = _err
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
