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

	"github.com/stretchr/testify/assert"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnConnect(func(sessionPresent bool) {
		assert.False(t, sessionPresent)

		done <- true
	})

	err := c.Connect("mqtt://localhost:1883", &Options{
		ClientID: "test",
	})

	assert.NoError(t, err)

	<-done

	c.Disconnect()
}

func TestClientConnectWebSocket(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnConnect(func(sessionPresent bool) {
		assert.False(t, sessionPresent)

		done <- true
	})

	err := c.Connect("ws://localhost:1884", &Options{
		ClientID: "test",
	})

	assert.NoError(t, err)

	<-done

	c.Disconnect()
}

func TestClientPublishSubscribe(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnMessage(func(topic string, payload []byte) {
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)

		done <- true
	})

	err := c.Connect("mqtt://localhost:1883", &Options{
		ClientID: "test",
	})

	assert.NoError(t, err)

	c.Subscribe("test", 0)
	c.Publish("test", []byte("test"), 0, false)

	<-done

	c.Disconnect()
}

func TestClientConnectError(t *testing.T) {
	c := NewClient()

	// wrong port
	err := c.Connect("mqtt://localhost:1234", &Options{
		ClientID: "test",
	})

	assert.Error(t, err)
}

func TestClientAuthenticationError(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnError(func(err error) {
		assert.Error(t, err)

		done <- true
	})

	// missing clientID
	err := c.Connect("mqtt://localhost:1883", &Options{})

	assert.NoError(t, err)

	<-done
}
