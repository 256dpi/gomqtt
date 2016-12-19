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
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/gomqtt/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWSServer(t *testing.T) {
	abstractServerTest(t, "ws")
}

func TestWSSServer(t *testing.T) {
	abstractServerTest(t, "wss")
}

func TestWSServerLaunchError(t *testing.T) {
	abstractServerLaunchErrorTest(t, "ws")
}

func TestWSSServerLaunchError(t *testing.T) {
	abstractServerLaunchErrorTest(t, "wss")
}

func TestWebSocketServerAcceptAfterClose(t *testing.T) {
	abstractServerAcceptAfterCloseTest(t, "ws")
}

func TestWebSocketServerCloseAfterClose(t *testing.T) {
	abstractServerCloseAfterClose(t, "ws")
}

func TestWebSocketServerInvalidUpgrade(t *testing.T) {
	port := tools.NewPort()

	server, err := testLauncher.Launch(port.URL("ws"))
	require.NoError(t, err)

	resp, err := http.PostForm(port.URL("http"), url.Values{"foo": {"bar"}})
	assert.Equal(t, "405 Method Not Allowed", resp.Status)
	assert.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func TestWebSocketServerAcceptAfterError(t *testing.T) {
	port := tools.NewPort()

	server, err := testLauncher.Launch(port.URL("ws"))
	require.NoError(t, err)

	webSocketServer := server.(*WebSocketServer)

	err = webSocketServer.listener.Close()
	assert.NoError(t, err)

	conn, err := server.Accept()
	require.Nil(t, conn)
	assert.Equal(t, NetworkError, toError(err).Code)
}

func TestWebSocketServerConnectionCancelOnClose(t *testing.T) {
	port := tools.NewPort()

	server, err := testLauncher.Launch(port.URL("ws"))
	require.NoError(t, err)

	conn, err := testDialer.Dial(port.URL("ws"))
	require.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, NetworkError, toError(err).Code)
}

func TestWebSocketFallback(t *testing.T) {
	port := tools.NewPort()

	server, err := testLauncher.Launch(port.URL("ws"))
	require.NoError(t, err)

	ws := server.(*WebSocketServer)

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})

	ws.SetFallback(mux)

	resp, err := http.Get(port.URL("http") + "/test")
	assert.NoError(t, err)
	assert.Equal(t, "200 OK", resp.Status)
	bytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello world!"), bytes)

	err = server.Close()
	assert.NoError(t, err)
}

func TestWebSocketOriginChecker(t *testing.T) {
	port := tools.NewPort()

	server, err := testLauncher.Launch(port.URL("ws"))
	require.NoError(t, err)

	ws := server.(*WebSocketServer)
	ws.SetOriginChecker(func(r *http.Request) bool {
		return false
	})

	conn, err := testDialer.Dial(port.URL("ws"))
	require.Error(t, err)
	require.Nil(t, conn)
}
