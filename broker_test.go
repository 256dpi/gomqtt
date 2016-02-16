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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/flow"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func abstractPublishSubscribeTest(t *testing.T, out, in string, sub, pub uint8) {
	port, done := startBroker(t, New(), 1)

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

	connectFuture, err := client.Connect(port.URL(), nil)
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
	port, done := startBroker(t, New(), 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), nil)
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

	connectFuture2, err := client2.Connect(port.URL(), nil)
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

func TestClearRetainedMessage(t *testing.T) {
	port, done := startBroker(t, New(), 3)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())

	publishFuture1, err := client1.Publish("test", []byte("test1"), 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture1.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test1"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())

	subscribeFuture1, err := client2.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())

	<-wait

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	client3 := client.New()
	client3.Callback = errorCallback(t)

	connectFuture3, err := client3.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())

	subscribeFuture2, err := client3.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())

	time.Sleep(50 * time.Millisecond)

	err = client3.Disconnect()
	assert.NoError(t, err)

	<-done
}

func abstractWillTest(t *testing.T, sub, pub uint8) {
	port, done := startBroker(t, New(), 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     pub,
	}

	connectFuture1, err := client1.Connect(port.URL(), opts)
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

	connectFuture2, err := client2.Connect(port.URL(), nil)
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
	port, done := startBroker(t, New(), 2)

	client1 := client.New()
	client1.Callback = errorCallback(t)

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     0,
		Retain:  true,
	}

	connectFuture1, err := client1.Connect(port.URL(), opts)
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

	connectFuture2, err := client2.Connect(port.URL(), nil)
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

func TestUnsubscribe(t *testing.T) {
	port, done := startBroker(t, New(), 1)

	client := client.New()
	client.Callback = errorCallback(t)

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subscribeFuture, err := client.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	unsubscribeFuture, err := client.Unsubscribe("test")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	time.Sleep(50 * time.Millisecond)

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestSubscriptionUpgrade(t *testing.T) {
	port, done := startBroker(t, New(), 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(1), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subscribeFuture1, err := client.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())

	subscribeFuture2, err := client.Subscribe("test", 1)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestMultipleSubscriptions(t *testing.T) {
	port, done := startBroker(t, New(), 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	subs := make(map[string]uint8)
	subs["test"] = 0
	subs["test"] = 1
	subs["test"] = 2

	subscribeFuture, err := client.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())

	publishFuture, err := client.Publish("test", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestConnectTimeout(t *testing.T) {
	broker := New()
	broker.ConnectTimeout = 10 * time.Millisecond

	port, done := startBroker(t, broker, 1)

	conn, err := transport.Dial(port.URL())
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func TestKeepAlive(t *testing.T) {
	t.Parallel()

	port, done := startBroker(t, New(), 1)

	opts := client.NewOptions()
	opts.KeepAlive = "1s"

	client := client.New()
	client.Callback = errorCallback(t)

	var reqCounter int32
	var respCounter int32

	client.Logger = func(message string) {
		if strings.Contains(message, "Pingreq") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "Pingresp") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	connectFuture, err := client.Connect(port.URL(), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())

	time.Sleep(2500 * time.Millisecond)

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done

	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(2), atomic.LoadInt32(&respCounter))
}

func TestKeepAliveTimeout(t *testing.T) {
	t.Parallel()

	connect := packet.NewConnectPacket()
	connect.KeepAlive = 1

	connack := packet.NewConnackPacket()

	client := flow.New().
		Send(connect).
		Receive(connack).
		End()

	port, done := startBroker(t, New(), 1)

	conn, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	client.Test(t, conn)

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
// connect without a clientId for MQTT 3.1.1
// disconnect another client with the same clientId
// disconnect if another broker connects the same client
// restore QoS 0 subscriptions not clean

// -- client pub sub
// offline message support for direct publish

// -- qos1
// restore QoS 1 subscriptions not clean
// remove stored subscriptions if connected with clean=true
// resend publish on non-clean reconnect QoS 1
// do not resend QoS 1 packets at each reconnect
// do not resend QoS 1 packets if reconnect is clean
// do not resend QoS 1 packets at reconnect if puback was received
// remove stored subscriptions after unsubscribe

// -- qos2
// restore QoS 2 subscriptions not clean
// resend publish on non-clean reconnect QoS 2
// resend pubrel on non-clean reconnect QoS 2
// publish after disconnection

// -- will
// delivers old will in case of a crash
