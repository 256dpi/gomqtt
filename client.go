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
	"net/url"
	"sync"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"gopkg.in/tomb.v2"
)

type (
	ConnectCallback func(bool)
	MessageCallback func(string, []byte)
	ErrorCallback   func(error)
)

type Client struct {
	opts *Options
	conn transport.Conn

	connectCallback ConnectCallback
	messageCallback MessageCallback
	errorCallback   ErrorCallback

	lastSend        time.Time
	lastSendMutex   sync.Mutex
	pingrespPending bool

	tomb tomb.Tomb
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{}
}

// OnConnect sets the callback for successful connections.
func (c *Client) OnConnect(callback ConnectCallback) {
	c.connectCallback = callback
}

// OnMessage sets the callback for incoming messages.
func (c *Client) OnMessage(callback MessageCallback) {
	c.messageCallback = callback
}

// OnError sets the callback for failed connection attempts and parsing errors.
func (c *Client) OnError(callback ErrorCallback) {
	c.errorCallback = callback
}

// Connect opens the connection to the broker.
func (c *Client) Connect(urlString string, opts *Options) error {
	var err error

	// parse url
	urlParts, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	// dial broker
	c.conn, err = transport.Dial(urlString)
	if err != nil {
		return err
	}

	// save opts
	if opts != nil {
		c.opts = opts
	} else {
		c.opts = NewOptions("gomqtt/client")
	}

	// preset to avoid pingreq right after start
	c.lastSend = time.Now()

	// start process routine
	c.tomb.Go(c.process)

	// start keep alive if greater than zero
	if c.opts.KeepAlive > 0 {
		c.tomb.Go(c.keepAlive)
	}

	// prepare connect packet
	m := packet.NewConnectPacket()
	m.ClientID = []byte(opts.ClientID)
	m.KeepAlive = uint16(opts.KeepAlive.Seconds())
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
	return c.send(m)
}

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

func (c *Client) Subscribe(topic string, qos byte) error {
	return c.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

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

func (c *Client) Unsubscribe(topic string) error {
	return c.UnsubscribeMultiple([]string{topic})
}

func (c *Client) UnsubscribeMultiple(topics []string) error {
	m := packet.NewUnsubscribePacket()
	m.Topics = make([][]byte, 0, len(topics))
	m.PacketID = 1

	for _, t := range topics {
		m.Topics = append(m.Topics, []byte(t))
	}

	return c.send(m)
}

func (c *Client) Disconnect() error {
	m := packet.NewDisconnectPacket()

	err := c.send(m)
	if err != nil {
		return err
	}

	return c.tomb.Wait()
}

// process incoming packets
func (c *Client) process() error {
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
			case packet.CONNACK:
				m, ok := pkt.(*packet.ConnackPacket)

				if ok {
					if m.ReturnCode == packet.ConnectionAccepted {
						if c.connectCallback != nil {
							c.connectCallback(m.SessionPresent)
						}
					} else {
						c.error(errors.New(m.ReturnCode.Error()))
					}
				} else {
					c.error(errors.New("failed to convert CONNACK packet"))
				}
			case packet.PINGRESP:
				_, ok := pkt.(*packet.PingrespPacket)

				if ok {
					c.pingrespPending = false
				} else {
					c.error(errors.New("failed to convert PINGRESP packet"))
				}
			case packet.PUBLISH:
				m, ok := pkt.(*packet.PublishPacket)

				if ok {
					c.messageCallback(string(m.Topic), m.Payload)
				} else {
					c.error(errors.New("failed to convert PUBLISH packet"))
				}
			}
		}
	}
}

func (c *Client) send(msg packet.Packet) error {
	c.lastSendMutex.Lock()
	c.lastSend = time.Now()
	c.lastSendMutex.Unlock()

	return c.conn.Send(msg)
}

func (c *Client) error(err error) {
	if c.errorCallback != nil {
		c.errorCallback(err)
	}
}

// manages the sending of ping packets to keep the connection alive
func (c *Client) keepAlive() error {
	for {
		c.lastSendMutex.Lock()
		timeElapsed := time.Since(c.lastSend)
		c.lastSendMutex.Unlock()

		timeToWait := c.opts.KeepAlive

		if timeElapsed > c.opts.KeepAlive {
			c.send(packet.NewPingreqPacket())
		} else {
			timeToWait = c.opts.KeepAlive - timeElapsed
		}

		select {
		case <-c.tomb.Dying():
			return tomb.ErrDying
		case <-time.After(timeToWait):
			// continue
		}
	}
}
