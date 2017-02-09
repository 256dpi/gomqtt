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
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func TestConnectTimeout(t *testing.T) {
	engine := NewEngine()
	engine.ConnectTimeout = 10 * time.Millisecond

	port, quit, done := Run(t, engine, "tcp")

	conn, err := transport.Dial("tcp://localhost:" + port)
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	close(quit)
	<-done
}

func TestDefaultReadLimit(t *testing.T) {
	engine := NewEngine()
	engine.DefaultReadLimit = 1

	port, quit, done := Run(t, engine, "tcp")

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) {
		assert.Error(t, err)
		close(wait)
	}

	cf, err := c.Connect(client.NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.Error(t, cf.Wait())

	<-wait
	close(quit)
	<-done
}
