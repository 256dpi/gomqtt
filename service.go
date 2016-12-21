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
	"github.com/jpillora/backoff"
	"gopkg.in/tomb.v2"
)

type command struct {
	publish     *publish
	subscribe   *subscribe
	unsubscribe *unsubscribe
}

type publish struct {
	msg    *packet.Message
	future *PublishFuture
}

type subscribe struct {
	subscriptions []packet.Subscription
	future        *SubscribeFuture
}

type unsubscribe struct {
	topics []string
	future *UnsubscribeFuture
}

// An OnlineCallback is a function that is called when the service is connected.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future will actually deadlock the service.
type OnlineCallback func(resumed bool)

// A MessageCallback is a function that is called when a message is received.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future will actually deadlock the service.
type MessageCallback func(*packet.Message)

// An ErrorCallback is a function that is called when an error occurred.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future will actually deadlock the service.
type ErrorCallback func(error)

// An OfflineCallback is a function that is called when the service is disconnected.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future will actually deadlock the service.
type OfflineCallback func()

const (
	serviceStarted byte = iota
	serviceStopped
)

// Service is an abstraction for Client that provides a stable interface to the
// application, while it automatically connects and reconnects clients in the
// background. Errors are not returned but logged using the Logger callback.
// All methods return Futures that get completed once the acknowledgements are
// received. Once the services is stopped all waiting futures get canceled.
//
// Note: If clean session is false and there are packets in the store, messages
// might get completed after starting without triggering any futures to complete.
type Service struct {
	config *Config

	state   *state
	backoff *backoff.Backoff

	// The session used by the client to store unacknowledged packets.
	Session Session

	// The callback that is used to notify that the service is online.
	OnlineCallback OnlineCallback

	// The callback to be called by the service upon receiving a message.
	MessageCallback MessageCallback

	// The callback to be called by the service upon encountering an error.
	ErrorCallback ErrorCallback

	// The callback that is used to notify that the service is offline.
	OfflineCallback OfflineCallback

	// The logger that is used to log write low level information like packets
	// that have ben successfully sent and received, details about the
	// automatic keep alive handler, reconnection and occurring errors.
	Logger Logger

	// The minimum delay between reconnects.
	//
	// Note: The value must be changed before calling Start.
	MinReconnectDelay time.Duration

	// The maximum delay between reconnects.
	//
	// Note: The value must be changed before calling Start.
	MaxReconnectDelay time.Duration

	// The allowed timeout until a connection attempt is canceled.
	ConnectTimeout time.Duration

	// The allowed timeout until a connection is forcefully closed.
	DisconnectTimeout time.Duration

	commandQueue chan *command
	futureStore  *futureStore

	mutex sync.Mutex
	tomb  *tomb.Tomb
}

// NewService allocates and returns a new service. The optional parameter queueSize
// specifies how many Subscribe, Unsubscribe and Publish commands can be queued
// up before actually sending them on the wire. The default queueSize is 100.
func NewService(queueSize ...int) *Service {
	var qs = 100
	if len(queueSize) > 0 {
		qs = queueSize[0]
	}

	return &Service{
		state:             newState(serviceStopped),
		Session:           NewMemorySession(),
		MinReconnectDelay: 1 * time.Second,
		MaxReconnectDelay: 32 * time.Second,
		ConnectTimeout:    5 * time.Second,
		DisconnectTimeout: 10 * time.Second,
		commandQueue:      make(chan *command, qs),
		futureStore:       newFutureStore(),
	}
}

// Start will start the service with the specified configuration. From now on
// the service will automatically reconnect on any error until Stop is called.
func (s *Service) Start(config *Config) {
	if config == nil {
		panic("No config specified")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// return if already started
	if s.state.get() == serviceStarted {
		return
	}

	// set state
	s.state.set(serviceStarted)

	// save config
	s.config = config

	// initialize backoff
	s.backoff = &backoff.Backoff{
		Min:    s.MinReconnectDelay,
		Max:    s.MaxReconnectDelay,
		Factor: 2,
	}

	// mark future store as protected
	s.futureStore.protect(true)

	// create new tomb
	s.tomb = &tomb.Tomb{}

	// start supervisor
	s.tomb.Go(s.supervisor)
}

// Publish will send a PublishPacket containing the passed parameters. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (s *Service) Publish(topic string, payload []byte, qos uint8, retain bool) *PublishFuture {
	msg := &packet.Message{
		Topic:   topic,
		Payload: payload,
		QOS:     qos,
		Retain:  retain,
	}

	return s.PublishMessage(msg)
}

// PublishMessage will send a PublishPacket containing the passed message. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (s *Service) PublishMessage(msg *packet.Message) *PublishFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// allocate future
	future := &PublishFuture{}
	future.initialize()

	// queue publish
	s.commandQueue <- &command{
		publish: &publish{
			msg:    msg,
			future: future,
		},
	}

	return future
}

// Subscribe will send a SubscribePacket containing one topic to subscribe. It
// will return a SubscribeFuture that gets completed once the acknowledgements
// have been received.
func (s *Service) Subscribe(topic string, qos uint8) *SubscribeFuture {
	return s.SubscribeMultiple([]packet.Subscription{
		{Topic: topic, QOS: qos},
	})
}

// SubscribeMultiple will send a SubscribePacket containing multiple topics to
// subscribe. It will return a SubscribeFuture that gets completed once the
// acknowledgements have been received.
func (s *Service) SubscribeMultiple(subscriptions []packet.Subscription) *SubscribeFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// allocate future
	future := &SubscribeFuture{}
	future.initialize()

	// queue subscribe
	s.commandQueue <- &command{
		subscribe: &subscribe{
			subscriptions: subscriptions,
			future:        future,
		},
	}

	return future
}

// Unsubscribe will send a UnsubscribePacket containing one topic to unsubscribe.
// It will return a SubscribeFuture that gets completed once the acknowledgements
// have been received.
func (s *Service) Unsubscribe(topic string) *UnsubscribeFuture {
	return s.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a UnsubscribePacket containing multiple
// topics to unsubscribe. It will return a SubscribeFuture that gets completed
// once the acknowledgements have been received.
func (s *Service) UnsubscribeMultiple(topics []string) *UnsubscribeFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// allocate future
	future := &UnsubscribeFuture{}
	future.initialize()

	// queue unsubscribe
	s.commandQueue <- &command{
		unsubscribe: &unsubscribe{
			topics: topics,
			future: future,
		},
	}

	return future
}

// Stop will disconnect the client if online and cancel all futures if requested.
// After the service is stopped in can be started again.
//
// Note: You should clear the futures on the last stop before exiting to ensure
// that all goroutines return that wait on futures.
func (s *Service) Stop(clearFutures bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// return if service not started
	if s.state.get() != serviceStarted {
		return
	}

	// kill and wait
	s.tomb.Kill(nil)
	s.tomb.Wait()

	// clear futures if requested
	if clearFutures {
		s.futureStore.protect(false)
		s.futureStore.clear()
	}

	// set state
	s.state.set(serviceStopped)
}

// the supervised reconnect loop
func (s *Service) supervisor() error {
	first := true

	for {
		if first {
			// no delay on first attempt
			first = false
		} else {
			// get backoff duration
			d := s.backoff.Duration()
			s.log(fmt.Sprintf("Delay Reconnect: %v", d))

			// sleep but return on Stop
			select {
			case <-time.After(d):
			case <-s.tomb.Dying():
				return tomb.ErrDying
			}
		}

		s.log("Next Reconnect")

		// prepare the stop channel
		fail := make(chan struct{})

		// try once to get a client
		client, resumed := s.connect(fail)
		if client == nil {
			continue
		}

		// run callback
		if s.OnlineCallback != nil {
			s.OnlineCallback(resumed)
		}

		// run dispatcher on client
		dying := s.dispatcher(client, fail)

		// run callback
		if s.OfflineCallback != nil {
			s.OfflineCallback()
		}

		// return goroutine if dying
		if dying {
			return tomb.ErrDying
		}
	}
}

// will try to connect one client to the broker
func (s *Service) connect(fail chan struct{}) (*Client, bool) {
	// prepare new client
	client := New()
	client.Session = s.Session
	client.Logger = s.Logger
	client.futureStore = s.futureStore

	// set callback
	client.Callback = func(msg *packet.Message, err error) {
		if err != nil {
			s.err("Client", err)
			close(fail)
			return
		}

		// call the handler
		if s.MessageCallback != nil {
			s.MessageCallback(msg)
		}
	}

	// attempt to connect
	future, err := client.Connect(s.config)
	if err != nil {
		s.err("Connect", err)
		return nil, false
	}

	// wait for connack
	err = future.Wait(s.ConnectTimeout)

	// check if future has been canceled
	if err == ErrFutureCanceled {
		s.err("Connect", err)
		return nil, false
	}

	// check if future has timed out
	if err == ErrFutureTimeout {
		client.Close()

		s.err("Connect", err)
		return nil, false
	}

	// check return code
	if future.ReturnCode != packet.ConnectionAccepted {
		client.Close()

		s.err("Connect", future.ReturnCode)
		return nil, false
	}

	return client, future.SessionPresent
}

// reads from the queues and calls the current client
func (s *Service) dispatcher(client *Client, fail chan struct{}) bool {
	for {
		select {
		case cmd := <-s.commandQueue:

			// handle subscribe command
			if cmd.subscribe != nil {
				future, err := client.SubscribeMultiple(cmd.subscribe.subscriptions)
				if err != nil {
					s.err("Subscribe", err)

					// cancel future
					cmd.subscribe.future.cancel()

					return false
				}

				cmd.subscribe.future.bind(future)
			}

			// handle unsubscribe command
			if cmd.unsubscribe != nil {
				future, err := client.UnsubscribeMultiple(cmd.unsubscribe.topics)
				if err != nil {
					s.err("Unsubscribe", err)

					// cancel future
					cmd.unsubscribe.future.cancel()

					return false
				}

				cmd.unsubscribe.future.bind(future)
			}

			// handle publish command
			if cmd.publish != nil {
				future, err := client.PublishMessage(cmd.publish.msg)
				if err != nil {
					s.err("Publish", err)
					return false
				}

				cmd.publish.future.bind(future)
			}
		case <-s.tomb.Dying():
			// disconnect client on Stop
			err := client.Disconnect(s.DisconnectTimeout)
			if err != nil {
				s.err("Disconnect", err)
			}

			return true
		case <-fail:
			return false
		}
	}
}

func (s *Service) err(sys string, err error) {
	s.log(fmt.Sprintf("%s Error: %s", sys, err.Error()))

	if s.ErrorCallback != nil {
		s.ErrorCallback(err)
	}
}

func (s *Service) log(str string) {
	if s.Logger != nil {
		s.Logger(str)
	}
}
