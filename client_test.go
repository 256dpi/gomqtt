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
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func testOptions() *Options {
	return NewOptions("test/" + uuid.NewV4().String())
}

func TestClientConnect(t *testing.T) {
	c := NewClient()

	future, err := c.Connect("mqtt://localhost:1883", testOptions())
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

	future, err := c.Connect("ws://localhost:1884", testOptions())
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientConnectAfterConnect(t *testing.T) {
	c := NewClient()

	future, err := c.Connect("mqtt://localhost:1883", testOptions())
	assert.NoError(t, err)
	assert.NoError(t, future.Wait())
	assert.False(t, future.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, future.ReturnCode)

	future, err = c.Connect("mqtt://localhost:1883", testOptions())
	assert.Equal(t, ErrAlreadyConnecting, err)
	assert.Nil(t, future)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func abstractPublishSubscribeTest(t *testing.T, qos byte) {
	c := NewClient()
	done := make(chan struct{})

	c.Callback = func(topic string, payload []byte, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)

		close(done)
	}

	connectFuture, err := c.Connect("mqtt://localhost:1883", testOptions())
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	subscribeFuture, err := c.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []byte{qos}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish("test", []byte("test"), qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-done
	err = c.Disconnect()
	assert.NoError(t, err)

	in, err := c.IncomingStore.All()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(in))

	out, err := c.OutgoingStore.All()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	abstractPublishSubscribeTest(t, 0)
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
	abstractPublishSubscribeTest(t, 1)
}

func TestClientPublishSubscribeQOS2(t *testing.T) {
	abstractPublishSubscribeTest(t, 2)
}

func TestClientUnsubscribe(t *testing.T) {
	c := NewClient()
	done := make(chan struct{})

	c.Callback = func(topic string, payload []byte, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)

		close(done)
	}

	connectFuture, err := c.Connect("mqtt://localhost:1883", testOptions())
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.False(t, connectFuture.SessionPresent)
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)

	subscribeFuture, err := c.Subscribe("foo", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []byte{0}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := c.Unsubscribe("foo")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	subscribeFuture, err = c.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []byte{0}, subscribeFuture.ReturnCodes)

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
	future, err := c.Connect("mqtt://localhost:1234", testOptions())
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

	c.Logger = func(message string) {
		if strings.Contains(message, "PingreqPacket") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "PingrespPacket") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	opts := testOptions()
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

// -- client
// should require a clientId with clean=false
// should have different client ids
// should delay closing everything up until the queue is depleted
// should delay ending up until all inflight messages are delivered
// does not wait acks when force-closing
// should reconnect if pingresp is not sent
// should not reconnect if pingresp is successful
// should defer the next ping when sending a control packet
// should fire a callback with error if disconnected (options provided)
// should fire a callback with error if disconnected (options not provided)
// should resend in-flight QoS 1 publish messages from the client
// should resend in-flight QoS 2 publish messages from the client

// -- store
// should put and stream in-flight packets
// should support destroying the stream
// should add and del in-flight packets
// should replace a packet when doing put with the same messageId
// should return the original packet on del
// should get a packet with the same messageId
