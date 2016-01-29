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

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	future, err := c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientConnectWebSocket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	c := NewClient()

	future, err := c.Connect("ws://localhost:1884", NewOptions("test"))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientConnectAfterConnect(t *testing.T) {
	c := NewClient()

	future, err := c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	future, err = c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.Equal(t, ErrAlreadyConnecting, err)
	assert.Nil(t, future)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func abstractPublishSubscribeTest(t *testing.T, qos byte) {
	c := NewClient()
	done := make(chan struct{})

	c.OnMessage(func(topic string, payload []byte) {
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)

		close(done)
	})

	connectFuture, err := c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	future, err := c.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())

	publishFuture, err := c.Publish("test", []byte("test"), qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-done
	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	abstractPublishSubscribeTest(t, 0)
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
	abstractPublishSubscribeTest(t, 1)
}

//func TestClientPublishSubscribeQOS2(t *testing.T) {
//	abstractPublishSubscribeTest(t, 2)
//}

func TestClientUnsubscribe(t *testing.T) {
	c := NewClient()
	done := make(chan struct{})

	c.OnMessage(func(topic string, payload []byte) {
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)

		close(done)
	})

	connectFuture, err := c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	future1, err := c.Subscribe("foo", 0)
	assert.NoError(t, err)
	assert.NoError(t, future1.Wait())

	future2, err := c.Unsubscribe("foo")
	assert.NoError(t, err)
	assert.NoError(t, future2.Wait())

	future1, err = c.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, future1.Wait())

	publishFuture, err := c.Publish("foo", []byte("test"), 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	publishFuture, err = c.Publish("test", []byte("test"), 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-done
	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientConnectError(t *testing.T) {
	c := NewClient()

	// wrong port
	future, err := c.Connect("mqtt://localhost:1234", NewOptions("test"))
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientAuthenticationError(t *testing.T) {
	c := NewClient()

	// missing clientID
	future, err := c.Connect("mqtt://localhost:1883", &Options{})
	assert.Error(t, err)
	assert.Nil(t, future)
}

func TestClientKeepAlive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	c := NewClient()

	var reqCounter int32 = 0
	var respCounter int32 = 0

	c.OnLog(func(message string) {
		if strings.Contains(message, "PingreqPacket") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "PingrespPacket") {
			atomic.AddInt32(&respCounter, 1)
		}
	})

	opts := NewOptions("test")
	opts.KeepAlive = "2s"

	future, err := c.Connect("mqtt://localhost:1883", opts)
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	<-time.After(7 * time.Second)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, int32(3), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(3), atomic.LoadInt32(&respCounter))
}

// -- abstract client
// should emit close if stream closes
// should mark the client as disconnected
// should stop ping timer if stream closes
// should emit close after end called
// should stop ping timer after end called
// should require a clientId with clean=false
// should emit connect
// should provide connack packet with connect event
// should mark the client as connected
// should emit error
// should have different client ids
// should queue message until connected
// should delay closing everything up until the queue is depleted
// should delay ending up until all inflight messages are delivered
// wait QoS 1 publish messages
// does not wait acks when force-closing
// should publish a message (offline)
// should publish a message (online)
// should accept options
// should fire a callback (qos 0)
// should fire a callback (qos 1)
// should fire a callback (qos 2)
// Publish 10 QoS 2 and receive them
// should send an unsubscribe packet (offline)
// should send an unsubscribe packet
// should accept an array of unsubs
// should fire a callback on unsuback
// should checkPing at keepalive interval
// should set a ping timer
// should not set a ping timer keepalive=0
// should reconnect if pingresp is not sent
// should not reconnect if pingresp is successful
// should defer the next ping when sending a control packet
// should send a subscribe message (offline)
// should send a subscribe message
// should accept an array of subscriptions
// should accept an hash of subscriptions
// should accept an options parameter
// should fire a callback on suback
// should fire a callback with error if disconnected (options provided)
// should fire a callback with error if disconnected (options not provided)
// should subscribe with a chinese topic
// should fire the message event
// should support binary data
// should emit a message event (qos=2)
// should emit a message event (qos=2) - repeated publish
// should support chinese topic
// should follow qos 0 semantics (trivial)
// should follow qos 1 semantics
// should follow qos 2 semantics
// should mark the client disconnecting if #end called
// should reconnect after stream disconnect
// should emit \'reconnect\' when reconnecting
// should emit \'offline\' after going offline
// should not reconnect if it was ended by the user
// should setup a reconnect timer on disconnect
// should allow specification of a reconnect period
// should resend in-flight QoS 1 publish messages from the client
// should resend in-flight QoS 2 publish messages from the client

// -- abstract core
// should put and stream in-flight packets
// should support destroying the stream
// should add and del in-flight packets
// should replace a packet when doing put with the same messageId
// should return the original packet on del
// should get a packet with the same messageId

// -- mqtt
// should throw an error when it is called with cert and key set but no protocol specified
// should throw an error when it is called with cert and key set and protocol other than allowed: mqtt,mqtts,ws,wss
// should return a MqttClient with mqtts set when connect is called key and cert set and protocol mqtt
// should return a MqttClient with mqtts set when connect is called key and cert set and protocol mqtts
// should return a MqttClient with wss set when connect is called key and cert set and protocol ws
// should return a MqttClient with wss set when connect is called key and cert set and protocol wss
// should support passing the key and cert
// should support passing the key, cert and CA list
// should fire callback on net connect
// should bind stream close to connection
// should bind stream error to conn
