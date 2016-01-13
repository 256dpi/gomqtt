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
	//"net"
	"testing"
	//"net/http"
	//"net/url"
	//"time"

	"github.com/stretchr/testify/require"
	"github.com/gomqtt/packet"
)

func abstractServerTest(t *testing.T, protocol string) {
	tp := newTestPort()

	server, err := testLauncher.Launch(tp.url(protocol))
	require.NoError(t, err)

	go func(){
		conn1, err := server.Accept()
		require.NoError(t, err)

		pkt, err := conn1.Receive()
		require.Equal(t, pkt.Type(), packet.CONNECT)
		require.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		require.NoError(t, err)

		pkt, err = conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	}()

	conn2, err := testDialer.Dial(tp.url(protocol))
	require.NoError(t, err)

	err = conn2.Send(packet.NewConnectPacket())
	require.NoError(t, err)

	pkt, err := conn2.Receive()
	require.Equal(t, pkt.Type(), packet.CONNACK)
	require.NoError(t, err)

	err = conn2.Close()
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)
}

func abstractServerLaunchErrorTest(t *testing.T, protocol string) {
	tp := testPort(1) // <- no permissions

	server, err := testLauncher.Launch(tp.url(protocol))
	require.Error(t, err)
	require.Nil(t, server)
}

//TODO: migrate
//func TestInactiveRequestHandler(t *testing.T) {
//	tp := newTestPort()
//
//	listener, err := net.Listen("tcp", tp.address())
//	require.NoError(t, err)
//
//	server := NewServer(noopHandler)
//	server.Stop()
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", server.RequestHandler())
//
//	serv := &http.Server{
//		Handler: mux,
//	}
//	go func(){
//		err = serv.Serve(listener)
//		require.NoError(t, err)
//	}()
//	_, err = testDialer.Dial(tp.url("ws"))
//	require.Error(t, err)
//
//	listener.Close()
//}

//TODO: migrate
//func TestInvalidWebSocketUpgrade(t *testing.T) {
//	tp := newTestPort()
//
//	server := NewServer(noopHandler)
//	server.Launch("ws", tp.address())
//
//	resp, err := http.PostForm(tp.url("http"), url.Values{"foo": {"bar"}})
//	require.Equal(t, "405 Method Not Allowed", resp.Status)
//	require.NoError(t, err)
//
//	server.Stop()
//	require.NoError(t, server.Error())
//}

//TODO: migrate
//func TestTCPAcceptError(t *testing.T) {
//	tp := newTestPort()
//
//	server := NewServer(noopHandler)
//	server.Launch("tcp", tp.address())
//
//	server.listeners[0].Close()
//	time.Sleep(1 * time.Second)
//
//	require.Error(t, server.Error())
//	require.True(t, server.Stopped())
//}

//TODO: migrate
//func TestHTTPAcceptError(t *testing.T) {
//	tp := newTestPort()
//
//	server := NewServer(noopHandler)
//	server.Launch("ws", tp.address())
//
//	server.listeners[0].Close()
//	time.Sleep(1 * time.Second)
//
//	require.Error(t, server.Error())
//	require.True(t, server.Stopped())
//}
