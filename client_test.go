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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func TestClientConnectWrongURL(t *testing.T) {
	c := New()
	c.Callback = errorCallback(t)

	// wrong url
	future, err := c.Connect(NewConfig("foo"))
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnectWrongKeepAlive(t *testing.T) {
	c := New()
	c.Callback = errorCallback(t)

	// wrong keep alive
	future, err := c.Connect(&Config{
		BrokerURL:    "mqtt://localhost:1234",
		KeepAlive:    "foo",
		CleanSession: true,
	})
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnectErrorWrongPort(t *testing.T) {
	c := New()
	c.Callback = errorCallback(t)

	// wrong port
	future, err := c.Connect(NewConfig("mqtt://localhost:1234"))
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnectErrorMissingClientID(t *testing.T) {
	c := New()
	c.Callback = errorCallback(t)

	// missing clientID when clean=false
	future, err := c.Connect(NewConfig("mqtt://localhost:1234"))
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientConnect(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientConnectCustomDialer(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	config := NewConfig("tcp://localhost:" + port)
	config.Dialer = transport.NewDialer()

	future, err := c.Connect(config)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientConnectAfterConnect(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	future, err = c.Connect(NewConfig("tcp://localhost:" + port))
	assert.Equal(t, ErrClientAlreadyConnecting, err)
	assert.Nil(t, future)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientConnectWithCredentials(t *testing.T) {
	connect := connectPacket()
	connect.Username = "test"
	connect.Password = "test"

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	future, err := c.Connect(NewConfig(fmt.Sprintf("tcp://test:test@localhost:%s/", port)))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done
}

func TestClientNotConnected(t *testing.T) {
	c := New()
	c.Callback = errorCallback(t)

	future1, err := c.Publish("test", []byte("test"), 0, false)
	assert.Nil(t, future1)
	assert.Equal(t, ErrClientNotConnected, err)

	future2, err := c.Subscribe("test", 0)
	assert.Nil(t, future2)
	assert.Equal(t, ErrClientNotConnected, err)

	future3, err := c.Unsubscribe("test")
	assert.Nil(t, future3)
	assert.Equal(t, ErrClientNotConnected, err)

	err = c.Disconnect()
	assert.Equal(t, ErrClientNotConnected, err)

	err = c.Close()
	assert.Equal(t, ErrClientNotConnected, err)
}

func TestClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = packet.ErrNotAuthorized

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.Nil(t, msg)
		assert.Equal(t, ErrClientConnectionDenied, err)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ErrNotAuthorized, future.ReturnCode)

	<-done
	<-wait
}

func TestClientExpectedConnack(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(packet.NewPingrespPacket()).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.Nil(t, msg)
		assert.Equal(t, ErrClientExpectedConnack, err)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.Equal(t, ErrFutureCanceled, future.Wait())

	<-done
	<-wait
}

func TestClientKeepAlive(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreqPacket()
	pingresp := packet.NewPingrespPacket()

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		Send(pingresp).
		Receive(pingreq).
		Send(pingresp).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	var reqCounter int32
	var respCounter int32

	c.Logger = func(message string) {
		if strings.Contains(message, "Pingreq") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "Pingresp") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	config := NewConfig("tcp://localhost:" + port)
	config.KeepAlive = "100ms"

	future, err := c.Connect(config)
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

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.Nil(t, msg)
		assert.Equal(t, ErrClientMissingPong, err)
		close(wait)
	}

	config := NewConfig("tcp://localhost:" + port)
	config.KeepAlive = "5ms"

	future, err := c.Connect(config)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-wait
	<-done
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []uint8{0}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	in, err := c.Session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(in))

	out, err := c.Session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 1}}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []uint8{1}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.PacketID = 2

	puback := packet.NewPubackPacket()
	puback.PacketID = 2

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(puback).
		Send(publish).
		Receive(puback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(1), msg.QOS)
		assert.False(t, msg.Retain)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", 1)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{1}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	in, err := c.Session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(in))

	out, err := c.Session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))
}

func TestClientPublishSubscribeQOS2(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 2}}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []uint8{2}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 2
	publish.PacketID = 2

	pubrec := packet.NewPubrecPacket()
	pubrec.PacketID = 2

	pubrel := packet.NewPubrelPacket()
	pubrel.PacketID = 2

	pubcomp := packet.NewPubcompPacket()
	pubcomp.PacketID = 2

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
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

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", 2)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{2}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	in, err := c.Session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(in))

	out, err := c.Session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))
}

func TestClientUnsubscribe(t *testing.T) {
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = []string{"test"}
	unsubscribe.PacketID = 1

	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = 1

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(unsubscribe).
		Send(unsuback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
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
	connect.ClientID = "test"
	connect.CleanSession = false

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.PacketID = 1

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	config := NewConfig("tcp://localhost:" + port)
	config.ClientID = "test"
	config.CleanSession = false

	connectFuture, err := c.Connect(config)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	publishFuture, err := c.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NotNil(t, publishFuture)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, ErrFutureCanceled, publishFuture.Wait())

	<-done

	list, err := c.Session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
}

func TestClientDisconnectWithTimeout(t *testing.T) {
	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.PacketID = 1

	puback := packet.NewPubackPacket()
	puback.PacketID = 1

	wait := func() {
		time.Sleep(100 * time.Millisecond)
	}

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Run(wait).
		Send(puback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	connectFuture, err := c.Connect(NewConfig("tcp://localhost:" + port))
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

	list, err := c.Session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))
}

func TestClientClose(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = errorCallback(t)

	connectFuture, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	err = c.Close()
	assert.NoError(t, err)

	<-done
}

func TestClientInvalidPackets(t *testing.T) {
	c := New()

	// state not connecting
	err := c.processConnack(packet.NewConnackPacket())
	assert.NoError(t, err)

	c.state.set(clientConnecting)

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
	connect.ClientID = "test"
	connect.CleanSession = false

	publish1 := packet.NewPublishPacket()
	publish1.Message.Topic = "test"
	publish1.Message.Payload = []byte("test")
	publish1.Message.QOS = 1
	publish1.PacketID = 1

	puback1 := packet.NewPubackPacket()
	puback1.PacketID = 1

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(publish1).
		Send(puback1).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Session.SavePacket(outgoing, publish1)
	c.Session.PacketID()
	c.Callback = errorCallback(t)

	config := NewConfig("tcp://localhost:" + port)
	config.ClientID = "test"
	config.CleanSession = false

	connectFuture, err := c.Connect(config)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	time.Sleep(20 * time.Millisecond)

	err = c.Disconnect()
	assert.NoError(t, err)

	<-done

	pkts, err := c.Session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pkts))
}

func TestClientUnexpectedClose(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.Nil(t, msg)
		assert.Error(t, err)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-wait
	<-done
}

func TestClientConnackFutureCancellation(t *testing.T) {
	broker := tools.NewFlow().
		Receive(connectPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.Nil(t, msg)
		assert.Error(t, err)
		close(wait)
	}

	future, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.Equal(t, ErrFutureCanceled, future.Wait())

	<-wait
	<-done
}

func TestClientFutureCancellation(t *testing.T) {
	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.PacketID = 1

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Close()

	done, port := fakeBroker(t, broker)

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		assert.Nil(t, msg)
		assert.Error(t, err)
	}

	connectFuture, err := c.Connect(NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	publishFuture, err := c.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.Equal(t, ErrFutureCanceled, publishFuture.Wait())

	<-done
}

func TestClientLogger(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []uint8{0}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	wait := make(chan struct{})

	c := New()
	c.Callback = func(msg *packet.Message, err error) {
		close(wait)
	}

	var counter uint32
	c.Logger = func(msg string) {
		atomic.AddUint32(&counter, 1)
	}

	future, _ := c.Connect(NewConfig("tcp://localhost:" + port))
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

func BenchmarkClientPublish(b *testing.B) {
	c := New()

	connectFuture, err := c.Connect(NewConfig("mqtt://0.0.0.0"))
	if err != nil {
		panic(err)
	}

	err = connectFuture.Wait()
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		_, err := c.Publish("test", []byte("test"), 0, false)
		if err != nil {
			panic(err)
		}
	}

	err = c.Disconnect()
	if err != nil {
		panic(err)
	}
}
