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
)

func TestWebSocketConnConnection(t *testing.T) {
	abstractConnConnectTest(t, "ws")
}

func TestWebSocketConnClose(t *testing.T) {
	abstractConnCloseTest(t, "ws")
}

func TestWebSocketConnEncodeError(t *testing.T) {
	abstractConnEncodeErrorTest(t, "ws")
}

func TestWebSocketConnDecode1Error(t *testing.T) {
	abstractConnDecodeError1Test(t, "ws")
}

func TestWebSocketConnDecode2Error(t *testing.T) {
	abstractConnDecodeError2Test(t, "ws")
}

func TestWebSocketConnDecode3Error(t *testing.T) {
	abstractConnDecodeError3Test(t, "ws")
}

func TestWebSocketConnSendAfterClose(t *testing.T) {
	abstractConnSendAfterCloseTest(t, "ws")
}

func TestWebSocketConnCounters(t *testing.T) {
	abstractConnCountersTest(t, "ws")
}

func TestWebSocketConnReadLimit(t *testing.T) {
	abstractConnReadLimitTest(t, "ws")
}

func TestWebSocketBadFrameError(t *testing.T) {
	conn2, done := abstractConnTestPreparer("ws", func(conn1 Conn) {
		buf := []byte{0x07, 0x00, 0x00, 0x00, 0x00} // <- bad frame

		if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.UnderlyingConn().Write(buf)
		} else {
			panic("not a websocket conn")
		}

		pkt, err := conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, NetworkError, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, NetworkError, toError(err).Code())

	<-done
}
