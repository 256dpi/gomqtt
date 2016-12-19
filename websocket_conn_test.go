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

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
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

func TestWebSocketConnCloseWhileSend(t *testing.T) {
	abstractConnCloseWhileSendTest(t, "ws")
}

func TestWebSocketSendAndCloseTest(t *testing.T) {
	abstractConnSendAndCloseTest(t, "ws")
}

func TestWebSocketConnCounters(t *testing.T) {
	abstractConnCountersTest(t, "ws")
}

func TestWebSocketConnReadLimit(t *testing.T) {
	abstractConnReadLimitTest(t, "ws")
}

func TestWebSocketConnReadTimeout(t *testing.T) {
	abstractConnReadTimeoutTest(t, "ws")
}

func TestWebSocketConnCloseAfterClose(t *testing.T) {
	abstractConnCloseAfterCloseTest(t, "ws")
}

func TestWebSocketConnAddr(t *testing.T) {
	abstractConnAddrTest(t, "ws")
}

func TestWebSocketConnBufferedSend(t *testing.T) {
	abstractConnBufferedSendTest(t, "ws")
}

func TestWebSocketConnBufferedSendAfterClose(t *testing.T) {
	abstractConnBufferedSendAfterCloseTest(t, "ws")
}

func TestWebSocketConnBigBufferedSendAfterClose(t *testing.T) {
	abstractConnBigBufferedSendAfterCloseTest(t, "ws")
}

func TestWebSocketBadFrameError(t *testing.T) {
	conn2, done := connectionPair("ws", func(conn1 Conn) {
		buf := []byte{0x07, 0x00, 0x00, 0x00, 0x00} // <- bad frame

		if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.UnderlyingConn().UnderlyingConn().Write(buf)
		} else {
			panic("not a websocket conn")
		}

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, NetworkError, toError(err).Code)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, NetworkError, toError(err).Code)

	<-done
}

func TestWebSocketChunkedMessage(t *testing.T) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "hello"
	pkt.Message.Payload = []byte("world")

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		buf := make([]byte, pkt.Len())
		pkt.Encode(buf)

		err := conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.BinaryMessage, buf[:7])
		assert.NoError(t, err)

		err = conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.BinaryMessage, buf[7:])
		assert.NoError(t, err)

		in, err := conn1.Receive()
		assert.Nil(t, in)
		assert.Equal(t, NetworkError, toError(err).Code)
	})

	in, err := conn2.Receive()
	assert.Nil(t, err)
	assert.Equal(t, pkt.String(), in.String())

	err = conn2.Close()
	assert.NoError(t, err)

	in, err = conn2.Receive()
	assert.Nil(t, in)
	assert.Equal(t, NetworkError, toError(err).Code)

	<-done
}

func TestWebSocketCoalescedMessage(t *testing.T) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "hello"
	pkt.Message.Payload = []byte("world")

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		buf := make([]byte, pkt.Len()*2)
		pkt.Encode(buf)
		pkt.Encode(buf[pkt.Len():])

		err := conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.BinaryMessage, buf)
		assert.NoError(t, err)

		in, err := conn1.Receive()
		assert.Nil(t, in)
		assert.Equal(t, NetworkError, toError(err).Code)
	})

	in, err := conn2.Receive()
	assert.Nil(t, err)
	assert.Equal(t, pkt.String(), in.String())

	in, err = conn2.Receive()
	assert.Nil(t, err)
	assert.Equal(t, pkt.String(), in.String())

	err = conn2.Close()
	assert.NoError(t, err)

	in, err = conn2.Receive()
	assert.Nil(t, in)
	assert.Equal(t, NetworkError, toError(err).Code)

	<-done
}

func BenchmarkWebSocketConn(b *testing.B) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "foo/bar/baz"

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		for i := 0; i < b.N; i++ {
			err := conn1.Send(pkt)
			if err != nil {
				panic(err)
			}
		}
	})

	for i := 0; i < b.N; i++ {
		_, err := conn2.Receive()
		if err != nil {
			panic(err)
		}
	}

	b.SetBytes(int64(pkt.Len() * 2))

	<-done
}

func BenchmarkWebSocketConnBuffered(b *testing.B) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "foo/bar/baz"

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		for i := 0; i < b.N; i++ {
			err := conn1.BufferedSend(pkt)
			if err != nil {
				panic(err)
			}
		}
	})

	for i := 0; i < b.N; i++ {
		_, err := conn2.Receive()
		if err != nil {
			panic(err)
		}
	}

	b.SetBytes(int64(pkt.Len() * 2))

	<-done
}

func TestWebSocketNotBinaryMessage(t *testing.T) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "hello"
	pkt.Message.Payload = []byte("world")

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		err := conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.TextMessage, []byte("hello"))
		assert.NoError(t, err)
	})

	in, err := conn2.Receive()
	assert.Equal(t, NetworkError, toError(err).Code)
	assert.Nil(t, in)

	<-done
}
