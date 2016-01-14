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
	"github.com/stretchr/testify/require"
	"net/http"
"net/url"
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

func TestInvalidWebSocketUpgrade(t *testing.T) {
	tp := newTestPort()

	server, err := testLauncher.Launch(tp.url("ws"))
	require.NoError(t, err)

	resp, err := http.PostForm(tp.url("http"), url.Values{"foo": {"bar"}})
	require.Equal(t, "405 Method Not Allowed", resp.Status)
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)
}
