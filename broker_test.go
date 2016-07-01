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
	"testing"
	"time"

	"github.com/gomqtt/spec"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func TestBrokerTCP(t *testing.T) {
	testBroker(t, "tcp")
}

func TestBrokerWS(t *testing.T) {
	testBroker(t, "ws")
}

func testBroker(t *testing.T, protocol string) {
	backend := NewMemoryBackend()
	backend.Logins = map[string]string{
		"allow": "allow",
	}

	port, done := Run(NewWithBackend(backend), protocol)

	config := spec.AllFeatures()
	config.URL = fmt.Sprintf("%s://allow:allow@localhost:%s", protocol, port.Port())
	config.DenyURL = fmt.Sprintf("%s://deny:deny@localhost:%s", protocol, port.Port())
	config.NoMessageWait = 50 * time.Millisecond
	config.MessageRetainWait = 50 * time.Millisecond

	spec.Run(t, config)

	close(done)
}

func TestConnectTimeout(t *testing.T) {
	broker := New()
	broker.ConnectTimeout = 10 * time.Millisecond

	port, done := Run(broker, "tcp")

	conn, err := transport.Dial(port.URL())
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	close(done)
}
