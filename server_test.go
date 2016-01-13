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
	"github.com/gomqtt/packet"
)

func abstractServerTest(t *testing.T, protocol string) {
	tp := newTestPort()

	server := NewServer(func(conn1 Conn){
		pkt, err := conn1.Receive()
		require.Equal(t, pkt.Type(), packet.CONNECT)
		require.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		require.NoError(t, err)

		pkt, err = conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	})

	server.TLSConfig = serverTLSConfig
	server.Launch(protocol, tp.address())

	conn2, err := testDialer.Dial(tp.url(protocol))
	require.NoError(t, err)

	err = conn2.Send(packet.NewConnectPacket())
	require.NoError(t, err)

	pkt, err := conn2.Receive()
	require.Equal(t, pkt.Type(), packet.CONNACK)
	require.NoError(t, err)

	err = conn2.Close()
	require.NoError(t, err)

	server.Stop()
	require.NoError(t, server.Error())
}

func TestTCPServer(t *testing.T) {
	abstractServerTest(t, "tcp")
}

func TestTLSServer(t *testing.T) {
	abstractServerTest(t, "tls")
}

func TestWSServer(t *testing.T) {
	abstractServerTest(t, "ws")
}

func TestWSSServer(t *testing.T) {
	abstractServerTest(t, "wss")
}

func TestServerNoError(t *testing.T) {
	server := NewServer(noopHandler)
	require.NoError(t, server.Error())
	server.Stop()
}

func TestTCPServerError1(t *testing.T) {
	server := NewServer(noopHandler)
	err := server.Launch("tcp", "localhost:1") // <- no permissions
	require.Error(t, err)
	server.Stop()
}

func TestTCPServerError2(t *testing.T) {
	server := NewServer(noopHandler)
	server.Stop()
	require.Error(t, server.Launch("tcp", "localhost:1"))
}

func TestTLSServerError1(t *testing.T) {
	server := NewServer(noopHandler)
	server.TLSConfig = serverTLSConfig
	err := server.Launch("tls", "localhost:1") // <- no permissions
	require.Error(t, err)
	server.Stop()
}

func TestTLSServerError2(t *testing.T) {
	server := NewServer(noopHandler)
	server.TLSConfig = serverTLSConfig
	server.Stop()
	require.Error(t, server.Launch("tls", "localhost:1"))
}

func TestWSServerError1(t *testing.T) {
	server := NewServer(noopHandler)
	err := server.Launch("ws", "localhost:1") // <- no permissions
	require.Error(t, err)
	server.Stop()
}

func TestWSServerError2(t *testing.T) {
	server := NewServer(noopHandler)
	server.Stop()
	require.Error(t, server.Launch("ws", "localhost:1"))
}

func TestWSSServerError1(t *testing.T) {
	server := NewServer(noopHandler)
	server.TLSConfig = serverTLSConfig
	err := server.Launch("wss", "localhost:1") // <- no permissions
	require.Error(t, err)
	server.Stop()
}

func TestWSSServerError2(t *testing.T) {
	server := NewServer(noopHandler)
	server.TLSConfig = serverTLSConfig
	server.Stop()
	require.Error(t, server.Launch("wss", "localhost:1"))
}

func TestInactiveRequestHandler(t *testing.T) {
	tp := newTestPort()

	listener, err := net.Listen("tcp", tp.address())
	require.NoError(t, err)

	server := NewServer(noopHandler)
	server.Stop()

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

	server := NewServer(noopHandler)
	server.Launch("ws", tp.address())

	resp, err := http.PostForm(tp.url("http"), url.Values{"foo": {"bar"}})
	require.Equal(t, "405 Method Not Allowed", resp.Status)
	require.NoError(t, err)

	server.Stop()
	require.NoError(t, server.Error())
}

func TestTCPAcceptError(t *testing.T) {
	tp := newTestPort()

	server := NewServer(noopHandler)
	server.Launch("tcp", tp.address())

	server.listeners[0].Close()
	time.Sleep(1 * time.Second)

	require.Error(t, server.Error())
	require.True(t, server.Stopped())
}

func TestHTTPAcceptError(t *testing.T) {
	tp := newTestPort()

	server := NewServer(noopHandler)
	server.Launch("ws", tp.address())

	server.listeners[0].Close()
	time.Sleep(1 * time.Second)

	require.Error(t, server.Error())
	require.True(t, server.Stopped())
}
