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

package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalDial(t *testing.T) {
	tp := newTestPort()

	server, err := testLauncher.Launch(tp.url("tcp"))
	assert.NoError(t, err)

	go func() {
		conn, err := server.Accept()
		assert.NoError(t, err)

		pkt, err := conn.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, NetworkError, toError(err).Code())
	}()

	conn, err := Dial(tp.url("tcp"))
	assert.NoError(t, err)

	err = conn.Close()
	assert.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func TestDialerBadURL(t *testing.T) {
	conn, err := Dial("foo")
	assert.Nil(t, conn)
	assert.Equal(t, DialError, toError(err).Code())
}

func TestDialerUnsupportedProtocol(t *testing.T) {
	conn, err := Dial("foo://localhost")
	assert.Nil(t, conn)
	assert.Equal(t, DialError, toError(err).Code())
	assert.Equal(t, ErrUnsupportedProtocol, toError(err).Err())
}

func TestDialerTCPError(t *testing.T) {
	conn, err := Dial("tcp://localhost:1234567")
	assert.Nil(t, conn)
	assert.Equal(t, DialError, toError(err).Code())
}

func TestDialerTLSError(t *testing.T) {
	conn, err := Dial("tls://localhost:1234567")
	assert.Nil(t, conn)
	assert.Equal(t, DialError, toError(err).Code())
}

func TestDialerWSError(t *testing.T) {
	conn, err := Dial("ws://localhost:1234567")
	assert.Nil(t, conn)
	assert.Equal(t, DialError, toError(err).Code())
}

func TestDialerWSSError(t *testing.T) {
	conn, err := Dial("wss://localhost:1234567")
	assert.Nil(t, conn)
	assert.Equal(t, DialError, toError(err).Code())
}

func abstractDefaultPortTest(t *testing.T, protocol string) {
	tp := newTestPort()

	server, err := testLauncher.Launch(tp.url(protocol))
	assert.NoError(t, err)

	go func() {
		conn, err := server.Accept()
		assert.NoError(t, err)

		pkt, err := conn.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, ConnectionClose, toError(err).Code())
	}()

	dialer := NewDialer()
	dialer.TLSConfig = clientTLSConfig
	dialer.DefaultTCPPort = tp.port()
	dialer.DefaultTLSPort = tp.port()
	dialer.DefaultWSPort = tp.port()
	dialer.DefaultWSSPort = tp.port()

	conn, err := dialer.Dial(protocol + "://localhost")
	assert.NoError(t, err)

	err = conn.Close()
	assert.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func TestTCPDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "tcp")
}

func TestTLSDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "tls")
}

func TestWSDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "ws")
}

func TestWSSDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "wss")
}
