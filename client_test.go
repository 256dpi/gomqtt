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
	"testing"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	"sync/atomic"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	sess, err := c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.NoError(t, err)
	assert.False(t, sess)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientConnectWebSocket(t *testing.T) {
	c := NewClient()

	sess, err := c.Connect("ws://localhost:1884", NewOptions("test"))
	assert.NoError(t, err)
	assert.False(t, sess)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientPublishSubscribe(t *testing.T) {
	c := NewClient()
	done := make(chan struct{})

	c.OnMessage(func(topic string, payload []byte) {
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)

		close(done)
	})

	sess, err := c.Connect("mqtt://localhost:1883", NewOptions("test"))
	assert.NoError(t, err)
	assert.False(t, sess)

	err = c.Subscribe("test", 0)
	assert.NoError(t, err)

	err = c.Publish("test", []byte("test"), 0, false)
	assert.NoError(t, err)

	<-done
	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestClientConnectError(t *testing.T) {
	c := NewClient()

	// wrong port
	sess, err := c.Connect("mqtt://localhost:1234", NewOptions("test"))
	assert.Error(t, err)
	assert.False(t, sess)
}

func TestClientAuthenticationError(t *testing.T) {
	c := NewClient()

	// missing clientID
	sess, err := c.Connect("mqtt://localhost:1883", &Options{})
	assert.Error(t, err)
	assert.False(t, sess)
}

func TestClientKeepAlive(t *testing.T) {
	c := NewClient()

	var reqCounter int32 = 0
	var respCounter int32 = 0

	c.OnLog(func(message string){
		if strings.Contains(message, "PingreqPacket") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "PingrespPacket") {
			atomic.AddInt32(&respCounter, 1)
		}
	})

	opts := NewOptions("test")
	opts.KeepAlive = "2s"

	sess, err := c.Connect("mqtt://localhost:1883", opts)
	assert.NoError(t, err)
	assert.False(t, sess)

	<-time.After(7 * time.Second)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, int32(3), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(3), atomic.LoadInt32(&respCounter))
}
