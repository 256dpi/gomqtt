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
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func TestBroker(t *testing.T) {
	AcceptanceTest(t, func() *Broker {
		return New()
	})
}

func TestUnsubscribe(t *testing.T) {
	port, done := runBroker(t, New(), 1)

	client := client.New()
	client.Callback = errorCallback(t)

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

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
	port, done := runBroker(t, New(), 1)

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
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture1, err := client.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture1.ReturnCodes)

	subscribeFuture2, err := client.Subscribe("test", 1)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())
	assert.Equal(t, []uint8{1}, subscribeFuture2.ReturnCodes)

	publishFuture, err := client.Publish("test", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestMultipleSubscriptions(t *testing.T) {
	port, done := runBroker(t, New(), 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test3", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: "test1", QOS: 0},
		{Topic: "test2", QOS: 1},
		{Topic: "test3", QOS: 2},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0, 1, 2}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish("test3", []byte("test"), 2, false)
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

	port, done := runBroker(t, broker, 1)

	conn, err := transport.Dial(port.URL())
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func TestKeepAlive(t *testing.T) {
	t.Parallel()

	port, done := runBroker(t, New(), 1)

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
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

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

	client := tools.NewFlow().
		Send(connect).
		Receive(connack).
		End()

	port, done := runBroker(t, New(), 1)

	conn, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	client.Test(t, conn)

	<-done
}

func TestConnectionDenied(t *testing.T) {
	backend := NewMemoryBackend()
	backend.Logins = make(map[string]string)
	backend.Logins["allow"] = "allow"

	broker := New()
	broker.Backend = backend

	port, done := runBroker(t, broker, 2)

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture1, err := client1.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ErrNotAuthorized, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	client2 := client.New()
	client2.Callback = errorCallback(t)

	url := fmt.Sprintf("tcp://allow:allow@localhost:%s/", port.Port())
	connectFuture2, err := client2.Connect(url, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func abstractStoredSubscriptionTest(t *testing.T, qos uint8) {
	port, done := runBroker(t, New(), 2)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(port.URL(), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.True(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish("test", []byte("test"), qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestStoredSubscriptionsQOS0(t *testing.T) {
	abstractStoredSubscriptionTest(t, 0)
}

func TestStoredSubscriptionsQOS1(t *testing.T) {
	abstractStoredSubscriptionTest(t, 1)
}

func TestStoredSubscriptionsQOS2(t *testing.T) {
	abstractStoredSubscriptionTest(t, 2)
}

func TestCleanStoredSubscriptions(t *testing.T) {
	port, done := runBroker(t, New(), 2)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	options.CleanSession = true

	client2 := client.New()
	client2.Callback = errorCallback(t)

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	time.Sleep(50 * time.Millisecond)

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestRemoveStoredSubscription(t *testing.T) {
	port, done := runBroker(t, New(), 2)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client1.Unsubscribe("test")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()
	client2.Callback = errorCallback(t)

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	time.Sleep(50 * time.Millisecond)

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestPublishResendQOS1(t *testing.T) {
	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = "test"

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test", QOS: 1},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = "test"
	publishOut.Message.QOS = 1

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = "test"
	publishIn.Message.QOS = 1

	pubackIn := packet.NewPubackPacket()
	pubackIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	port, done := runBroker(t, New(), 2)

	conn1, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Send(subscribe).
		Skip(). // suback
		Send(publishOut).
		Skip(). // puback
		Receive(publishIn).
		Close().
		Test(t, conn1)

	conn2, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(publishIn).
		Send(pubackIn).
		Send(disconnect).
		Close().
		Test(t, conn2)

	<-done
}

func TestPubrelResendQOS2(t *testing.T) {
	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = "test"

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test", QOS: 2},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = "test"
	publishOut.Message.QOS = 2

	pubrelOut := packet.NewPubrelPacket()
	pubrelOut.PacketID = 2

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = "test"
	publishIn.Message.QOS = 2

	pubrecIn := packet.NewPubrecPacket()
	pubrecIn.PacketID = 1

	pubrelIn := packet.NewPubrelPacket()
	pubrelIn.PacketID = 1

	pubcompIn := packet.NewPubcompPacket()
	pubcompIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	port, done := runBroker(t, New(), 2)

	conn1, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Send(subscribe).
		Skip(). // suback
		Send(publishOut).
		Skip(). // pubrec
		Send(pubrelOut).
		Skip(). // pubcomp
		Receive(publishIn).
		Send(pubrecIn).
		Close().
		Test(t, conn1)

	conn2, err := transport.Dial(port.URL())
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(pubrelIn).
		Send(pubcompIn).
		Send(disconnect).
		Close().
		Test(t, conn2)

	<-done
}

func TestOfflineMessages(t *testing.T) {
	port, done := runBroker(t, New(), 3)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	/* offline subscriber */

	client1 := client.New()
	client1.Callback = errorCallback(t)

	connectFuture1, err := client1.Connect(port.URL(), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", 2)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{2}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	/* publisher */

	client2 := client.New()
	client2.Callback = errorCallback(t)

	connectFuture2, err := client2.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish("test", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(port.URL(), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	assert.NoError(t, err)

	<-done
}

// disconnect another client with the same clientId
// failed authentication does not disconnect other client with same clientId
// delivers old will in case of a crash
