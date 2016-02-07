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
		End()

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
		End()

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
		End()

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

func TestClientConnectionDenied(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ErrNotAuthorized)).
		Close()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		assert.Equal(t, ErrConnectionDenied, err)
		assert.Empty(t, topic)
		assert.Nil(t, payload)
		close(wait)
	}

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ErrNotAuthorized, future.ReturnCode)

	<-done
	<-wait
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
		Receive(disconnectPacket()).
		End()

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

	opts := NewOptions()
	opts.KeepAlive = "100ms"

	future, err := c.Connect(tp.url("tcp"), opts)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-time.After(250 * time.Millisecond)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(2), atomic.LoadInt32(&respCounter))

	<-done
}

func TestClientKeepAliveTimeout(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreqPacket()

	broker := flow.New().
		Receive(connect).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(pingreq).
		End()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		assert.Empty(t, topic)
		assert.Nil(t, payload)
		assert.Equal(t, ErrMissingPong, err)
		close(wait)
	}

	opts := NewOptions()
	opts.KeepAlive = "5ms"

	future, err := c.Connect(tp.url("tcp"), opts)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-wait
	<-done
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: []byte("test")},
	}
	subscribe.PacketID = 0

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{0}
	suback.PacketID = 0

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
		End()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)
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
	subscribe.PacketID = 0

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{1}
	suback.PacketID = 0

	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")
	publish.QOS = 1
	publish.PacketID = 1

	puback := packet.NewPubackPacket()
	puback.PacketID = 1

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
		End()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)
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

func TestClientPublishSubscribeQOS2(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: []byte("test"), QOS: 2},
	}
	subscribe.PacketID = 0

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{2}
	suback.PacketID = 0

	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")
	publish.QOS = 2
	publish.PacketID = 1

	pubrec := packet.NewPubrecPacket()
	pubrec.PacketID = 1

	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = 1

	pubcomp := packet.NewPubcompPacket()
	pubcomp.PacketID = 1

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(pubrec).
		Receive(pubrel).
		Send(pubcomp).
		Send(publish).
		Receive(pubrec).
		Send(pubrel).
		Receive(pubcomp).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)
		close(wait)
	}

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", 2)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []byte{2}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), 2, false)
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

func TestClientUnsubscribe(t *testing.T) {
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = [][]byte{[]byte("test")}
	unsubscribe.PacketID = 0

	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = 0

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(unsubscribe).
		Send(unsuback).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	unsubscribeFuture, err := c.Unsubscribe("test")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientHardDisconnect(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = []byte("test")
	connect.CleanSession = false

	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")
	publish.QOS = 1
	publish.PacketID = 0

	broker := flow.New().
		Receive(connect).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	opts := NewOptions()
	opts.ClientID = "test"
	opts.CleanSession = false

	connectFuture, err := c.Connect(tp.url("tcp"), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	publishFuture, err := c.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NotNil(t, publishFuture)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, ErrCanceled, publishFuture.Wait())

	<-done

	pkts, err := c.Session.AllPackets(session.Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pkts))
}

func TestClientDisconnectWithTimeout(t *testing.T) {
	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")
	publish.QOS = 1
	publish.PacketID = 0

	puback := packet.NewPubackPacket()
	puback.PacketID = 0

	wait := func() {
		time.Sleep(100 * time.Millisecond)
	}

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(publish).
		Run(wait).
		Send(puback).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	connectFuture, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	publishFuture, err := c.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NotNil(t, publishFuture)

	err = c.Disconnect(10 * time.Second)
	assert.NoError(t, err)

	<-done

	assert.NoError(t, publishFuture.Wait())

	pkts, err := c.Session.AllPackets(session.Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pkts))
}

func TestClientClose(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		End()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Callback = errorCallback(t)

	connectFuture, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	err = c.Close()
	assert.NoError(t, err)

	<-done
}

func TestClientInvalidPackets(t *testing.T) {
	c := NewClient()

	// state not connecting
	err := c.processConnack(packet.NewConnackPacket())
	assert.NoError(t, err)

	c.state.set(stateConnecting)

	// missing future
	err = c.processSuback(packet.NewSubackPacket())
	assert.NoError(t, err)

	// missing future
	err = c.processUnsuback(packet.NewUnsubackPacket())
	assert.NoError(t, err)

	// missing future
	err = c.processPubrel(0)
	assert.NoError(t, err)

	// missing future
	err = c.processPubackAndPubcomp(0)
	assert.NoError(t, err)
}

func TestClientSessionResumption(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = []byte("test")
	connect.CleanSession = false

	publish1 := packet.NewPublishPacket()
	publish1.Topic = []byte("test")
	publish1.Payload = []byte("test")
	publish1.QOS = 1
	publish1.PacketID = 0

	puback1 := packet.NewPubackPacket()
	puback1.PacketID = 0

	broker := flow.New().
		Receive(connect).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(publish1).
		Send(puback1).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	c := NewClient()
	c.Session.SavePacket(session.Outgoing, publish1)
	c.Session.PacketID()
	c.Callback = errorCallback(t)

	opts := NewOptions()
	opts.ClientID = "test"
	opts.CleanSession = false

	connectFuture, err := c.Connect(tp.url("tcp"), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	time.Sleep(20 * time.Millisecond)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	pkts, err := c.Session.AllPackets(session.Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pkts))
}

func TestClientUnexpectedClose(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket(packet.ConnectionAccepted)).
		Close()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		assert.Equal(t, ErrUnexpectedClose, err)
		assert.Empty(t, topic)
		assert.Nil(t, payload)
		close(wait)
	}

	future, err := c.Connect(tp.url("tcp"), nil)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-wait
	<-done
}

func TestClientLogger(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: []byte("test")},
	}
	subscribe.PacketID = 0

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{0}
	suback.PacketID = 0

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
		End()

	done, tp := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := NewClient()
	c.Callback = func(topic string, payload []byte, err error) {
		close(wait)
	}

	var counter uint32
	c.Logger = func(msg string) {
		atomic.AddUint32(&counter, 1)
	}

	future, _ := c.Connect(tp.url("tcp"), nil)
	future.Wait()

	subscribeFuture, _ := c.Subscribe("test", 0)
	subscribeFuture.Wait()

	publishFuture, _ := c.Publish("test", []byte("test"), 0, false)
	publishFuture.Wait()

	<-wait

	c.Disconnect()

	<-done

	assert.Equal(t, uint32(8), counter)
}

//func TestClientStoreError1(t *testing.T) {
//	c := NewClient()
//	c.Session = &testSession{ resetError: true }
//
//	connectFuture, err := c.Connect("mqtt://localhost:1883", testOptions())
//	assert.Error(t, err)
//	assert.Nil(t, connectFuture)
//}
