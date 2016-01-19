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
	"net/url"
	"sync"
	"time"
	"fmt"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"gopkg.in/tomb.v2"
)

type (
	MessageCallback func(string, []byte)
	ErrorCallback   func(error)
	LogCallback     func(string)
)

type Client struct {
	conn transport.Conn

	messageCallback MessageCallback
	errorCallback   ErrorCallback
	logCallback     LogCallback

	keepAlive       time.Duration
	lastSend        time.Time
	lastSendMutex   sync.Mutex
	pingrespPending bool

	tomb tomb.Tomb
	boot sync.WaitGroup
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{}
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
func (c* Client) OnLog(callback LogCallback) {
	c.logCallback = callback
}

// Connect opens the connection to the broker and sends a ConnectPacket.
func (c *Client) Connect(urlString string, opts *Options) (bool, error) {
	// parse url
	urlParts, err := url.Parse(urlString)
	if err != nil {
		return false, err
	}

	// dial broker
	c.conn, err = transport.Dial(urlString)
	if err != nil {
		return false, err
	}

	// save opts
	if opts == nil {
		opts = NewOptions("gomqtt/client")
	}

	// parse keep alive
	c.keepAlive, err = time.ParseDuration(opts.KeepAlive)
	if err != nil {
		return false, err
	}

	// prepare connect packet
	m := packet.NewConnectPacket()
	m.ClientID = []byte(opts.ClientID)
	m.KeepAlive = uint16(c.keepAlive.Seconds())
	m.CleanSession = opts.CleanSession

	// check for credentials
	if urlParts.User != nil {
		m.Username = []byte(urlParts.User.Username())
		p, _ := urlParts.User.Password()
		m.Password = []byte(p)
	}

	// set will
	m.WillTopic = []byte(opts.WillTopic)
	m.WillPayload = opts.WillPayload
	m.WillQOS = opts.WillQos
	m.WillRetain = opts.WillRetained

	// send connect packet
	err = c.send(m)
	if err != nil {
		return false, err
	}

	// receive connack packet
	pkt, err := c.conn.Receive()
	if err != nil {
		return false, err
	}

	// hold session present flag
	var sessionPresent bool

	// check packet type
	switch pkt.Type() {
	case packet.CONNACK:
		connack := pkt.(*packet.ConnackPacket)

		if connack.ReturnCode != packet.ConnectionAccepted {
			return false, connack.ReturnCode
		}

		sessionPresent = connack.SessionPresent
	}

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

	return sessionPresent, nil
}

// Publish will send a PublishPacket containing the passed parameters.
func (c *Client) Publish(topic string, payload []byte, qos byte, retain bool) error {
	m := packet.NewPublishPacket()
	m.Topic = []byte(topic)
	m.Payload = payload
	m.QOS = qos
	m.Retain = retain
	m.Dup = false
	m.PacketID = 1

	return c.send(m)
}

// Subscribe will send a SubscribePacket containing one topic to subscribe.
func (c *Client) Subscribe(topic string, qos byte) error {
	return c.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

// SubscribeMultiple will send a SubscribePacket containing multiple topics to
// subscribe.
func (c *Client) SubscribeMultiple(filters map[string]byte) error {
	m := packet.NewSubscribePacket()
	m.Subscriptions = make([]packet.Subscription, 0, len(filters))
	m.PacketID = 1

	for topic, qos := range filters {
		m.Subscriptions = append(m.Subscriptions, packet.Subscription{
			Topic: []byte(topic),
			QOS:   qos,
		})
	}

	return c.send(m)
}

// Unsubscribe will send a UnsubscribePacket containing one topic to unsubscribe.
func (c *Client) Unsubscribe(topic string) error {
	return c.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a UnsubscribePacket containing multiple
// topics to unsubscribe.
func (c *Client) UnsubscribeMultiple(topics []string) error {
	m := packet.NewUnsubscribePacket()
	m.Topics = make([][]byte, 0, len(topics))
	m.PacketID = 1

	for _, t := range topics {
		m.Topics = append(m.Topics, []byte(t))
	}

	return c.send(m)
}

// Disconnect will send a DisconnectPacket and close the connection.
func (c *Client) Disconnect() error {
	m := packet.NewDisconnectPacket()

	err := c.send(m)
	if err != nil {
		_err, ok := err.(transport.Error)

		if ok && _err.Code() != transport.ExpectedClose {
			return err
		}
	}

	c.tomb.Wait()

	return nil
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
				return err
			}

			switch pkt.Type() {
			case packet.PINGRESP:
				c.pingrespPending = false
				c.log("Received PingrespPacket")
			case packet.PUBLISH:
				publish := pkt.(*packet.PublishPacket)
				go c.messageCallback(string(publish.Topic), publish.Payload)
			default:
				c.log(fmt.Sprintf("Unhandled Packet: %s", pkt.Type().String()))
			}
		}
	}
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
		go c.errorCallback(err)
	}
}

// calls the log callback
func (c *Client) log(message string) {
	if c.logCallback != nil {
		go c.logCallback(message)
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
				return err
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
