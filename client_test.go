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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomqtt/flow"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/session"
	"github.com/stretchr/testify/assert"
)

func TestClientConnectError1(t *testing.T) {
	c := NewClient()
	c.Callback = errorCallback(t)

	// wrong url
	future, err := c.Connect("foo", nil)
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnectError2(t *testing.T) {
	c := NewClient()
	c.Callback = errorCallback(t)

	// wrong keep alive
	future, err := c.Connect("mqtt://localhost:1234", &Options{
		KeepAlive: "foo", CleanSession: true,
	})
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnectError3(t *testing.T) {
	c := NewClient()
	c.Callback = errorCallback(t)

	// wrong port
	future, err := c.Connect("mqtt://localhost:1234", nil)
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnectError4(t *testing.T) {
	c := NewClient()
	c.Callback = errorCallback(t)

	// missing clientID when clean=false
	future, err := c.Connect("mqtt://localhost:1234", &Options{})
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnect(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(disconnectPacket()).
		Close()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientConnectAfterConnect(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(disconnectPacket()).
		Close()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	future, err = c.Connect(tp.url("tcp"), nil)
	assert.Equal(t, ErrAlreadyConnecting, err)
	assert.Nil(t, future)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientConnectWithCredentials(t *testing.T) {
	connect := connectPacket()
	connect.Username = []byte("test")
	connect.Password = []byte("test")

	broker := flow.New().
		Receive(connect).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(disconnectPacket()).
		Close()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	future, err := c.Connect(tp.protectedURL("tcp", "test", "test"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientNotConnected(t *testing.T) {
	c := NewClient()
	c.Callback = errorCallback(t)

	future1, err := c.Publish("test", []byte("test"), 0, false)
	assert.Nil(t, future1)
	assert.Equal(t, ErrNotConnected, err)

	future2, err := c.Subscribe("test", 0)
	assert.Nil(t, future2)
	assert.Equal(t, ErrNotConnected, err)

	future3, err := c.Unsubscribe("test")
	assert.Nil(t, future3)
	assert.Equal(t, ErrNotConnected, err)

	err = c.Disconnect()
	assert.Equal(t, ErrNotConnected, err)
}

func TestClientKeepAlive(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreqPacket()
	pingresp := packet.NewPingrespPacket()

	broker := flow.New().
		Receive(connect).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(pingreq).
		Send(pingresp).
		Receive(pingreq).
		Send(pingresp).
		Receive(pingreq).
		Send(pingresp).
		Receive(disconnectPacket()).
		Close()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	var reqCounter int32
	var respCounter int32

	c.Logger = func(message string) {
		if strings.Contains(message, "PINGREQ") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "PINGRESP") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	opts := NewOptions("gomqtt/client")
	opts.KeepAlive = "100ms"

	future, err := c.Connect(tp.url("tcp"), opts)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-time.After(350 * time.Millisecond)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, int32(3), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(3), atomic.LoadInt32(&respCounter))

	<-done
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: []byte("test")},
	}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{0}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		Close()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(msg *Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		close(wait)
	}

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []byte{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	in, err := c.Session.AllPackets(session.Incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(in))

	out, err := c.Session.AllPackets(session.Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: []byte("test"), QOS: 1},
	}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{1}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")
	publish.QOS = 1
	publish.PacketID = 2

	puback := packet.NewPubackPacket()
	puback.PacketID = 2

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(puback).
		Send(publish).
		Receive(puback).
		Receive(disconnectPacket()).
		Close()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(msg *Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		close(wait)
	}

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", 1)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []byte{1}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	in, err := c.Session.AllPackets(session.Incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(in))

	out, err := c.Session.AllPackets(session.Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))
}

//func TestClientUnsubscribe(t *testing.T) {
//	c := NewClient()
//	c.Callback = errorCallback(t)
//	done := make(chan struct{})
//
//	c.Callback = func(msg *Message, err error) {
//		assert.NoError(t, err)
//		assert.Equal(t, "test", msg.Topic)
//		assert.Equal(t, []byte("test"), msg.Payload)
//
//		close(done)
//	}
//
//	connectFuture, err := c.Connect("mqtt://localhost:1883", testOptions())
//	assert.NoError(t, err)
//	assert.NoError(t, connectFuture.Wait())
//	assert.False(t, connectFuture.SessionPresent)
//	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
//
//	subscribeFuture, err := c.Subscribe("foo", 0)
//	assert.NoError(t, err)
//	assert.NoError(t, subscribeFuture.Wait())
//	assert.Equal(t, []byte{0}, subscribeFuture.ReturnCodes)
//
//	unsubscribeFuture, err := c.Unsubscribe("foo")
//	assert.NoError(t, err)
//	assert.NoError(t, unsubscribeFuture.Wait())
//
//	subscribeFuture, err = c.Subscribe("test", 0)
//	assert.NoError(t, err)
//	assert.NoError(t, subscribeFuture.Wait())
//	assert.Equal(t, []byte{0}, subscribeFuture.ReturnCodes)
//
//	publishFuture, err := c.Publish("foo", []byte("test"), 0, false)
//	assert.NoError(t, err)
//	assert.NoError(t, publishFuture.Wait())
//
//	publishFuture, err = c.Publish("test", []byte("test"), 0, false)
//	assert.NoError(t, err)
//	assert.NoError(t, publishFuture.Wait())
//
//	<-done
//	err = c.Disconnect()
//	assert.NoError(t, err)
//}
//
//func TestClientDisconnectWithTimeout(t *testing.T) {
//	c := NewClient()
//	c.Callback = errorCallback(t)
//
//	connectFuture, err := c.Connect("mqtt://localhost:1883", testOptions())
//	assert.NoError(t, err)
//	assert.NoError(t, connectFuture.Wait())
//	assert.False(t, connectFuture.SessionPresent)
//	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
//
//	for i:=0; i<10; i++ {
//		publishFuture, err := c.Publish("test", []byte("test"), 2, false)
//		assert.NoError(t, err)
//		assert.NotNil(t, publishFuture)
//	}
//
//	err = c.Disconnect(10 * time.Second)
//	assert.NoError(t, err)
//
//	pkts, err := c.Session.AllPackets(session.Outgoing)
//	assert.NoError(t, err)
//	assert.Equal(t, 0, len(pkts))
//}
//
//func TestClientInvalidPackets(t *testing.T) {
//	c := NewClient()
//
//	// state not connecting
//	err := c.processConnack(packet.NewConnackPacket())
//	assert.NoError(t, err)
//
//	c.state.set(stateConnecting)
//
//	err = c.processConnack(packet.NewConnackPacket())
//	assert.NoError(t, err)
//
//	err = c.processSuback(packet.NewSubackPacket())
//	assert.NoError(t, err)
//
//	err = c.processUnsuback(packet.NewUnsubackPacket())
//	assert.NoError(t, err)
//
//	err = c.processPubackAndPubcomp(0)
//	assert.NoError(t, err)
//}
//
//func TestClientStoreError1(t *testing.T) {
//	c := NewClient()
//	c.Session = &testSession{ resetError: true }
//
//	connectFuture, err := c.Connect("mqtt://localhost:1883", testOptions())
//	assert.Error(t, err)
//	assert.Nil(t, connectFuture)
//}
