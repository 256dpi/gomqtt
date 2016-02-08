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
	"fmt"
	"sync"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/session"
	"github.com/jpillora/backoff"
)

// ClearSession will connect/disconnect once with a clean session request to force
// the broker to reset the clients session. This is useful in situations where
// its not clear in what state the last session was left.
func ClearSession(url string, clientID string) error {
	client := NewClient()

	// prepare options
	options := NewOptions()
	options.ClientID = clientID
	options.CleanSession = true

	// connect to broker
	future, err := client.Connect(url, options)
	if err != nil {
		return err
	}

	// wait for connack
	future.Wait()

	// check if connection has been accepted
	if future.ReturnCode != packet.ConnectionAccepted {
		return ErrConnectionDenied
	}

	// disconnect
	return client.Disconnect()
}

type publish struct {
	topic   string
	payload []byte
	qos     byte
	retain  bool
	future  *PublishFuture
}

type subscribe struct {
	filters map[string]byte
	future  *SubscribeFuture
}

type unsubscribe struct {
	topics []string
	future *UnsubscribeFuture
}

type Handler func(topic string, payload []byte)

type Notifier func(online bool, sessionPresent bool)

type Service struct {
	broker  string
	options *Options
	backoff *backoff.Backoff

	Session  session.Session
	Handler  Handler
	Notifier Notifier
	Logger   Logger

	MinReconnectDelay time.Duration
	MaxReconnectDelay time.Duration
	ConnectTimeout    time.Duration
	DisconnectTimeout time.Duration

	subscribeQueue   chan *subscribe
	unsubscribeQueue chan *unsubscribe
	publishQueue     chan *publish
	stopChannel      chan time.Duration

	started bool
	mutex   sync.Mutex
}

func NewService() *Service {
	return &Service{
		Session:           session.NewMemorySession(),
		MinReconnectDelay: 1 * time.Second,
		MaxReconnectDelay: 32 * time.Second,
		ConnectTimeout:    5 * time.Second,
		DisconnectTimeout: 10 * time.Second,
		subscribeQueue:    make(chan *subscribe, 100),
		unsubscribeQueue:  make(chan *unsubscribe, 100),
		publishQueue:      make(chan *publish, 100),
		stopChannel:       make(chan time.Duration),
	}
}

func (s *Service) Start(url string, opts *Options) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return
	}

	s.broker = url
	s.options = opts

	s.backoff = &backoff.Backoff{
		Min:    s.MinReconnectDelay,
		Max:    s.MaxReconnectDelay,
		Factor: 2,
	}

	s.started = true

	go s.connect()
}

// Publish will send a PublishPacket containing the passed parameters. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (s *Service) Publish(topic string, payload []byte, qos byte, retain bool) *PublishFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	future := &PublishFuture{}
	future.initialize()

	s.publishQueue <- &publish{
		topic:   topic,
		payload: payload,
		qos:     qos,
		retain:  retain,
		future:  future,
	}

	return future
}

// Subscribe will send a SubscribePacket containing one topic to subscribe.
func (s *Service) Subscribe(topic string, qos byte) *SubscribeFuture {
	return s.SubscribeMultiple(map[string]byte{
		topic: qos,
	})
}

// SubscribeMultiple will send a SubscribePacket containing multiple topics to
// subscribe.
func (s *Service) SubscribeMultiple(filters map[string]byte) *SubscribeFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	future := &SubscribeFuture{}
	future.initialize()

	s.subscribeQueue <- &subscribe{
		filters: filters,
		future:  future,
	}

	return future
}

// Unsubscribe will send a UnsubscribePacket containing one topic to unsubscribe.
func (s *Service) Unsubscribe(topic string) *UnsubscribeFuture {
	return s.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a UnsubscribePacket containing multiple
// topics to unsubscribe.
func (s *Service) UnsubscribeMultiple(topics []string) *UnsubscribeFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	future := &UnsubscribeFuture{}
	future.initialize()

	s.unsubscribeQueue <- &unsubscribe{
		topics: topics,
		future: future,
	}

	return future
}

func (s *Service) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.started {
		return
	}

	s.started = false

	s.stopChannel <- s.DisconnectTimeout
}

func (s *Service) connect() {
	client := NewClient()
	client.Session = s.Session
	client.Callback = s.callback
	client.Logger = s.Logger

	s.log("Reconnect")

	future, err := client.Connect(s.broker, s.options)
	if err != nil {
		s.log("Connect Error: %v", err)
		s.reconnect()
		return
	}

	err = future.Wait(s.ConnectTimeout)
	if err == ErrCanceled {
		s.log("Connack: %v", err)
		s.reconnect()
		return
	} else if err == ErrTimeoutExceeded {
		client.Close()

		s.log("Connack: %v", err)
		s.reconnect()
		return
	}

	if future.ReturnCode != packet.ConnectionAccepted {
		client.Close()

		s.log("Connack: %s", future.ReturnCode.Error())
		s.reconnect()
		return
	}

	s.notify(true, future.SessionPresent)

	s.dispatcher(client)
}

func (s *Service) reconnect() {
	d := s.backoff.Duration()
	s.log("Delay Reconnect: %v", d)

	// TODO: break on Stop()
	time.Sleep(d)

	s.connect()
}

// reads from the queues and calls the current client
func (s *Service) dispatcher(client *Client) {
Loop:
	for {
		select {
		case sub := <-s.subscribeQueue:
			future, err := client.SubscribeMultiple(sub.filters)
			if err != nil {
				//TODO: requeue subscribe?
				s.log("Subscribe Error: %v", err)
				break Loop
			}

			sub.future.bind(future)
		case unsub := <-s.unsubscribeQueue:
			future, err := client.UnsubscribeMultiple(unsub.topics)
			if err != nil {
				//TODO: requeue unsubscribe?
				s.log("Unsubscribe Error: %v", err)
				break Loop
			}

			unsub.future.bind(future)
		case msg := <-s.publishQueue:
			future, err := client.Publish(msg.topic, msg.payload, msg.qos, msg.retain)
			if err != nil {
				s.log("Publish Error: %v", err)
				break Loop
			}

			msg.future.bind(future)
		case timeout := <-s.stopChannel:
			err := client.Disconnect(timeout)
			if err != nil {
				s.log("Disconnect Error: %v", err)
			}

			// TODO: Retain the clients future store

			break Loop
		}
	}

	s.notify(false, false)
}

func (s *Service) callback(topic string, payload []byte, err error) {
	if err != nil {
		s.log("Error: %v", err)

		// begin reconnect
		go s.reconnect()

		return
	}

	// call the handler
	if s.Handler != nil {
		s.Handler(topic, payload)
	}
}

func (s *Service) notify(online bool, resumed bool) {
	if s.Notifier != nil {
		s.Notifier(online, resumed)
	}
}

// log a message
func (s *Service) log(format string, a ...interface{}) {
	if s.Logger != nil {
		s.Logger(fmt.Sprintf(format, a...))
	}
}
