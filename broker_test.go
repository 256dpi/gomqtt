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

package broker

import (
	"testing"

	"github.com/gomqtt/client"
	"github.com/stretchr/testify/assert"
	"github.com/gomqtt/packet"
)

func TestPublishSubscribeQOS0(t *testing.T) {
	tp, done := startBroker(t, New(), 1)

	client := client.New()

	received := false
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)

		received = true
		close(wait)
	}

	connectFuture, err := client.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subscribeFuture, err := client.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done

	assert.True(t, received)
}

func TestPublishSubscribeQOS1(t *testing.T) {
	tp, done := startBroker(t, New(), 1)

	client := client.New()

	received := false
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)

		received = true
		close(wait)
	}

	connectFuture, err := client.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subscribeFuture, err := client.Subscribe("test", 1)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done

	assert.True(t, received)
}

func TestPublishSubscribeQOS2(t *testing.T) {
	tp, done := startBroker(t, New(), 1)

	client := client.New()

	received := false
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)

		received = true
		close(wait)
	}

	connectFuture, err := client.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subscribeFuture, err := client.Subscribe("test", 2)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done

	assert.True(t, received)
}

// -- authentication
// authenticate successfully a client with username and password
// authenticate unsuccessfully a client with username and password
// authenticate errors
// authorize publish
// do not authorize publish
// authorize subscribe
// negate subscription
// failed authentication does not disconnect other client with same clientId

// -- basic
// connect and connack (minimal)
// does not die badly on connection error
// unsubscribe
// unsubscribe on disconnect
// disconnect
// retain messages
// closes
// connect without a clientId for MQTT 3.1.1
// disconnect another client with the same clientId
// disconnect if another broker connects the same client
// publish to $SYS/broker/new/clients
// restore QoS 0 subscriptions not clean
// double sub does not double deliver

// -- client pub sub
// publish direct to a single client QoS 0
// publish direct to a single client QoS 1
// offline message support for direct publish
// subscribe a client programmatically
// subscribe a client programmatically multiple topics
// subscribe a client programmatically with full packet
// handle multiple subscriptions with the same filter
// respect the max wos set by a subscription

// -- keepalive
// supports pingreq/pingresp
// supports keep alive disconnections
// supports keep alive disconnections after a pingreq
// disconnect if a connect does not arrive in time

// -- qos1
// publish QoS 1
// subscribe QoS 1
// subscribe QoS 0, but publish QoS 1
// restore QoS 1 subscriptions not clean
// remove stored subscriptions if connected with clean=true
// resend publish on non-clean reconnect QoS 1
// do not resend QoS 1 packets at each reconnect
// do not resend QoS 1 packets if reconnect is clean
// do not resend QoS 1 packets at reconnect if puback was received
// deliver QoS 1 retained messages
// deliver QoS 0 retained message with QoS 1 subscription
// remove stored subscriptions after unsubscribe
// upgrade a QoS 0 subscription to QoS 1

// -- qos2
// publish QoS 2
// subscribe QoS 2
// subscribe QoS 0, but publish QoS 2
// restore QoS 2 subscriptions not clean
// resend publish on non-clean reconnect QoS 2
// resend pubrel on non-clean reconnect QoS 2
// publish after disconnection

// -- error
// after an error, outstanding packets are discarded

// -- will
// delivers a will
// delivers old will in case of a crash
// store the will in the persistence
