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

func abstractPublishSubscribeTest(t *testing.T, out, in string, sub, pub uint8) {
	tp, done := startBroker(t, New(), 1)

	client := client.New()

	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subscribeFuture, err := client.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	publishFuture, err := client.Publish(out, []byte("test"), pub, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestPublishSubscribeQOS0(t *testing.T) {
	abstractPublishSubscribeTest(t, "test", "test", 0, 0)
}

func TestPublishSubscribeQOS1(t *testing.T) {
	abstractPublishSubscribeTest(t, "test", "test", 1, 1)
}

func TestPublishSubscribeQOS2(t *testing.T) {
	abstractPublishSubscribeTest(t, "test", "test", 2, 2)
}

func TestPublishSubscribeWildcardOne(t *testing.T) {
	abstractPublishSubscribeTest(t, "foo/bar", "foo/+", 0, 0)
}

func TestPublishSubscribeWildcardSome(t *testing.T) {
	abstractPublishSubscribeTest(t, "foo/bar", "#", 0, 0)
}

func TestPublishSubscribeDowngrade1(t *testing.T) {
	abstractPublishSubscribeTest(t, "test", "test", 0, 1)
}

func TestPublishSubscribeDowngrade2(t *testing.T) {
	abstractPublishSubscribeTest(t, "test", "test", 0, 2)
}

func TestPublishSubscribeDowngrade3(t *testing.T) {
	abstractPublishSubscribeTest(t, "test", "test", 1, 2)
}

func abstractRetainedMessageTest(t *testing.T, out, in string, sub, pub uint8) {
	tp, done := startBroker(t, New(), 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())

	publishFuture, err := client1.Publish(out, []byte("test"), pub, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())

	subscribeFuture, err := client2.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestRetainedMessageQOS0(t *testing.T) {
	abstractRetainedMessageTest(t, "test", "test", 0, 0)
}

func TestRetainedMessageQOS1(t *testing.T) {
	abstractRetainedMessageTest(t, "test", "test", 1, 1)
}

func TestRetainedMessageQOS2(t *testing.T) {
	abstractRetainedMessageTest(t, "test", "test", 2, 2)
}

func TestRetainedMessageWildcardOne(t *testing.T) {
	abstractRetainedMessageTest(t, "foo/bar", "foo/+", 0, 0)
}

func TestRetainedMessageWildcardSome(t *testing.T) {
	abstractRetainedMessageTest(t, "foo/bar", "#", 0, 0)
}

func abstractWillTest(t *testing.T, sub, pub uint8) {
	tp, done := startBroker(t, New(), 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic: "test",
		Payload: []byte("test"),
		QOS: pub,
	}

	connectFuture1, err := client1.Connect(tp.url(), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())

	subscribeFuture, err := client2.Subscribe("test", sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	err = client1.Close()
	assert.NoError(t, err)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestWillQOS0(t *testing.T) {
	abstractWillTest(t, 0, 0)
}

func TestWillQOS1(t *testing.T) {
	abstractWillTest(t, 1, 1)
}

func TestWillQOS2(t *testing.T) {
	abstractWillTest(t, 2, 2)
}

func TestRetainedWill(t *testing.T) {
	tp, done := startBroker(t, New(), 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic: "test",
		Payload: []byte("test"),
		QOS: 0,
		Retain: true,
	}

	connectFuture1, err := client1.Connect(tp.url(), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())

	err = client1.Close()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(tp.url(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())

	subscribeFuture, err := client2.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
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
// does not die badly on connection error
// unsubscribe
// unsubscribe on disconnect
// disconnect
// closes
// connect without a clientId for MQTT 3.1.1
// disconnect another client with the same clientId
// disconnect if another broker connects the same client
// restore QoS 0 subscriptions not clean
// double sub does not double deliver

// -- client pub sub
// offline message support for direct publish
// handle multiple subscriptions with the same filter

// -- keepalive
// supports pingreq/pingresp
// supports keep alive disconnections
// supports keep alive disconnections after a pingreq
// disconnect if a connect does not arrive in time

// -- qos1
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
// restore QoS 2 subscriptions not clean
// resend publish on non-clean reconnect QoS 2
// resend pubrel on non-clean reconnect QoS 2
// publish after disconnection

// -- retained message
// clear retained message

// -- error
// after an error, outstanding packets are discarded

// -- will
// delivers old will in case of a crash
