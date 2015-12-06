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
	"github.com/gomqtt/stream"
)

type (
	ConnectCallback func(bool)
	MessageCallback func(string, []byte)
	ErrorCallback   func(error)
)

type Client struct {
	opts   *Options
	stream stream.Stream
	quit   chan bool

	connectCallback ConnectCallback
	messageCallback MessageCallback
	errorCallback   ErrorCallback

	lastContact     time.Time
	pingrespPending bool
	start           sync.WaitGroup
}

// NewClient returns a new client.
func NewClient() *Client {
	return &Client{
		quit: make(chan bool),
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

// QuickConnect opens a connection to the broker specified by an URL.
func (this *Client) QuickConnect(urlString string, clientID string) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	opts := &Options{
		URL:          url,
		ClientID:     clientID,
		CleanSession: true,
	}

	return this.Connect(opts)
}

// Connect opens the connection to the broker.
func (this *Client) Connect(opts *Options) error {
	var err error

	// dial broker
	this.stream, err = dial(opts)
	if err != nil {
		return err
	}

	// save opts
	this.opts = opts

	// start watcher and processor
	this.start.Add(3)
	go this.process()
	go this.keepAlive()
	go this.watch()

	// prepare connect packet
	m := packet.NewConnectPacket()
	m.ClientID = []byte(opts.ClientID)
	m.KeepAlive = uint16(opts.KeepAlive.Seconds())
	m.CleanSession = opts.CleanSession

	// check for credentials
	if opts.URL.User != nil {
		m.Username = []byte(opts.URL.User.Username())
		p, _ := opts.URL.User.Password()
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
}

// process incoming packets
func (this *Client) process() {
	this.start.Done()

	for {
		select {
		case <-this.quit:
			return
		case msg, ok := <-this.stream.Incoming():
			if !ok {
				return
			}

			this.lastContact = time.Now()

			switch msg.Type() {
			case packet.CONNACK:
				m, ok := msg.(*packet.ConnackPacket)

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
				_, ok := msg.(*packet.PingrespPacket)

				if ok {
					this.pingrespPending = false
				} else {
					this.error(errors.New("failed to convert PINGRESP packet"))
				}
			case packet.PUBLISH:
				m, ok := msg.(*packet.PublishPacket)

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
	this.lastContact = time.Now()
	this.stream.Send(msg)
}

func (this *Client) error(err error) {
	if this.errorCallback != nil {
		this.errorCallback(err)
	}
}

// manages the sending of ping packets to keep the connection alive
func (this *Client) keepAlive() {
	this.start.Done()

	for {
		select {
		case <-this.quit:
			return
		default:
			last := uint(time.Since(this.lastContact).Seconds())

			if last > uint(this.opts.KeepAlive.Seconds()) {
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

// watch for EOFs and errors on the stream
func (this *Client) watch() {
	this.start.Done()

	for {
		select {
		case <-this.quit:
			return
		case err, ok := <-this.stream.Error():
			if !ok {
				return
			}

			this.error(err)
		}
	}
}
