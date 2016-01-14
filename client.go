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
)

type (
	ConnectCallback func(bool)
	MessageCallback func(string, []byte)
	ErrorCallback   func(error)
)

type Client struct {
	opts *Options
	conn transport.Conn
	quit chan bool

	connectCallback ConnectCallback
	messageCallback MessageCallback
	errorCallback   ErrorCallback

	lastContact      time.Time
	lastContactMutex sync.Mutex
	pingrespPending  bool
	start            sync.WaitGroup
	finish			 sync.WaitGroup
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{
		quit: make(chan bool),
		lastContact: time.Now(),
	}
}

// OnConnect sets the callback for successful connections.
func (this *Client) OnConnect(callback ConnectCallback) {
	this.connectCallback = callback
}

// OnMessage sets the callback for incoming messages.
func (this *Client) OnMessage(callback MessageCallback) {
	this.messageCallback = callback
}

// OnError sets the callback for failed connection attempts and parsing errors.
func (this *Client) OnError(callback ErrorCallback) {
	this.errorCallback = callback
}

// Connect opens the connection to the broker.
func (this *Client) Connect(urlString string, opts *Options) error {
	var err error

	// parse url
	urlParts, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	// dial broker
	this.conn, err = transport.Dial(urlString)
	if err != nil {
		return err
	}

	// save opts
	if opts != nil {
		this.opts = opts
	} else {
		this.opts = &Options{
			ClientID:     "gomqtt/client",
			CleanSession: true,
		}
	}

	// start subroutines
	this.start.Add(2)
	this.finish.Add(2)
	go this.process()
	go this.keepAlive()

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
	this.send(m)

	this.start.Wait()
	return nil
}

func (this *Client) Publish(topic string, payload []byte, qos byte, retain bool) {
	m := packet.NewPublishPacket()
	m.Topic = []byte(topic)
	m.Payload = payload
	m.QOS = qos
	m.Retain = retain
	m.Dup = false
	m.PacketID = 1

	this.send(m)
}

func (this *Client) Subscribe(topic string, qos byte) {
	this.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

func (this *Client) SubscribeMultiple(filters map[string]byte) {
	m := packet.NewSubscribePacket()
	m.Subscriptions = make([]packet.Subscription, 0, len(filters))
	m.PacketID = 1

	for topic, qos := range filters {
		m.Subscriptions = append(m.Subscriptions, packet.Subscription{
			Topic: []byte(topic),
			QOS:   qos,
		})
	}

	this.send(m)
}

func (this *Client) Unsubscribe(topic string) {
	this.UnsubscribeMultiple([]string{topic})
}

func (this *Client) UnsubscribeMultiple(topics []string) {
	m := packet.NewUnsubscribePacket()
	m.Topics = make([][]byte, 0, len(topics))
	m.PacketID = 1

	for _, t := range topics {
		m.Topics = append(m.Topics, []byte(t))
	}

	this.send(m)
}

func (this *Client) Disconnect() {
	m := packet.NewDisconnectPacket()

	this.send(m)

	close(this.quit)

	this.finish.Wait()
}

// process incoming packets
func (this *Client) process() {
	this.start.Done()
	defer this.finish.Done()

	for {
		select {
		case <-this.quit:
			return
		default:
			pkt, err := this.conn.Receive()
			if err != nil {
				//TODO: handle error
				return
			}

			//this.lastContact = time.Now()

			switch pkt.Type() {
			case packet.CONNACK:
				m, ok := pkt.(*packet.ConnackPacket)

				if ok {
					if m.ReturnCode == packet.ConnectionAccepted {
						if this.connectCallback != nil {
							this.connectCallback(m.SessionPresent)
						}
					} else {
						this.error(errors.New(m.ReturnCode.Error()))
					}
				} else {
					this.error(errors.New("failed to convert CONNACK packet"))
				}
			case packet.PINGRESP:
				_, ok := pkt.(*packet.PingrespPacket)

				if ok {
					this.pingrespPending = false
				} else {
					this.error(errors.New("failed to convert PINGRESP packet"))
				}
			case packet.PUBLISH:
				m, ok := pkt.(*packet.PublishPacket)

				if ok {
					this.messageCallback(string(m.Topic), m.Payload)
				} else {
					this.error(errors.New("failed to convert PUBLISH packet"))
				}
			}
		}
	}
}

func (this *Client) send(msg packet.Packet) {
	this.lastContactMutex.Lock()
	this.lastContact = time.Now()
	this.lastContactMutex.Unlock()

	this.conn.Send(msg)
}

func (this *Client) error(err error) {
	if this.errorCallback != nil {
		this.errorCallback(err)
	}
}

// manages the sending of ping packets to keep the connection alive
func (this *Client) keepAlive() {
	this.start.Done()
	defer this.finish.Done()

	for {
		select {
		case <-this.quit:
			return
		default:
			this.lastContactMutex.Lock()
			last := uint(time.Since(this.lastContact).Seconds())
			this.lastContactMutex.Unlock()

			if this.opts.KeepAlive.Seconds() > 0 && last > uint(this.opts.KeepAlive.Seconds()) {
				if !this.pingrespPending {
					this.pingrespPending = true

					m := packet.NewPingreqPacket()
					this.send(m)
				} else {
					// TODO: close connection?
					// return
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}
