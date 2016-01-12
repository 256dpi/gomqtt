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
	"net"
	"testing"
	"net/http"
	"net/url"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerNoError(t *testing.T) {
	server := NewServer(make(chan Conn))
	require.NoError(t, server.Error())
}

func TestTCPServerError1(t *testing.T) {
	server := NewServer(make(chan Conn))
	err := server.LaunchTCP("localhost:1") // <- no permissions
	require.Error(t, err)
}

func TestTCPServerError2(t *testing.T) {
	server := NewServer(make(chan Conn))
	server.tomb.Kill(nil) // <- fake stopped server
	require.Error(t, server.LaunchTCP("localhost:1"))
}

func TestTLSServerError1(t *testing.T) {
	server := NewServer(make(chan Conn))
	err := server.LaunchTLS("localhost:1", serverTLSConfig) // <- no permissions
	require.Error(t, err)
}

func TestTLSServerError2(t *testing.T) {
	server := NewServer(make(chan Conn))
	server.tomb.Kill(nil) // <- fake stopped server
	require.Error(t, server.LaunchTLS("localhost:1", serverTLSConfig))
}

func TestWSServerError1(t *testing.T) {
	server := NewServer(make(chan Conn))
	err := server.LaunchWS("localhost:1") // <- no permissions
	require.Error(t, err)
}

func TestWSServerError2(t *testing.T) {
	server := NewServer(make(chan Conn))
	server.tomb.Kill(nil) // <- fake stopped server
	require.Error(t, server.LaunchWS("localhost:1"))
}

func TestWSSServerError1(t *testing.T) {
	server := NewServer(make(chan Conn))
	err := server.LaunchWSS("localhost:1", serverTLSConfig) // <- no permissions
	require.Error(t, err)
}

func TestWSSServerError2(t *testing.T) {
	server := NewServer(make(chan Conn))
	server.tomb.Kill(nil) // <- fake stopped server
	require.Error(t, server.LaunchWSS("localhost:1", serverTLSConfig))
}

func TestInactiveRequestHandler(t *testing.T) {
	tp := newTestPort()

	listener, err := net.Listen("tcp", tp.address())
	require.NoError(t, err)

	server := NewServer(make(chan Conn))
	server.tomb.Kill(nil) // <- fake stopped server

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.RequestHandler())

	serv := &http.Server{
		Handler: mux,
	}
	go func(){
		err = serv.Serve(listener)
		require.NoError(t, err)
	}()
	_, err = testDialer.Dial(tp.url("ws"))
	require.Error(t, err)

	listener.Close()
}

func TestInvalidWebSocketUpgrade(t *testing.T) {
	tp := newTestPort()

	server := NewServer(make(chan Conn))
	server.LaunchWS(tp.address())

	resp, err := http.PostForm(tp.url("http"), url.Values{"foo": {"bar"}})
	require.Equal(t, "405 Method Not Allowed", resp.Status)
	require.NoError(t, err)

	server.Stop()
	require.NoError(t, server.Error())
}

func TestTCPServerStop(t *testing.T) {
	tp := newTestPort()

	server := NewServer(make(chan Conn))
	server.LaunchTCP(tp.address())

	conn, err := testDialer.Dial(tp.url("tcp"))
	require.NoError(t, err)

	server.Stop()
	require.NoError(t, server.Error())
	require.True(t, server.Stopped())

	_, err = conn.Receive()
	require.Error(t, err)
}

func TestWSServerStop(t *testing.T) {
	tp := newTestPort()

	server := NewServer(make(chan Conn))
	server.LaunchWS(tp.address())

	conn, err := testDialer.Dial(tp.url("ws"))
	require.NoError(t, err)

	server.Stop()
	require.NoError(t, server.Error())
	require.True(t, server.Stopped())

	_, err = conn.Receive()
	require.Error(t, err)
}

func TestTCPAcceptError(t *testing.T) {
	tp := newTestPort()

	server := NewServer(make(chan Conn))
	server.LaunchTCP(tp.address())

	server.listeners[0].Close()
	time.Sleep(1 * time.Second)

	require.Error(t, server.Error())
	require.True(t, server.Stopped())
}

func TestHTTPAcceptError(t *testing.T) {
	tp := newTestPort()

	server := NewServer(make(chan Conn))
	server.LaunchWS(tp.address())

	server.listeners[0].Close()
	time.Sleep(1 * time.Second)

	require.Error(t, server.Error())
	require.True(t, server.Stopped())
}
