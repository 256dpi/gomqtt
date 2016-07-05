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
	"errors"
	"testing"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestIsConnectionCloseError(t *testing.T) {
	assert.True(t, IsConnectionCloseError(newTransportError(ConnectionClose, nil)))
	assert.False(t, IsConnectionCloseError(nil))
	assert.False(t, IsConnectionCloseError(errors.New("foo")))
}

func abstractConnConnectTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, ConnectionClose, toError(err).Code())
	})

	err := conn2.Send(packet.NewConnectPacket())
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	<-done
}

func abstractConnCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnEncodeErrorTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt := packet.NewConnackPacket()
		pkt.ReturnCode = 11 // <- invalid return code

		err := conn1.Send(pkt)
		assert.Equal(t, EncodeError, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnDecodeError1Test(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		buf := []byte{0x00, 0x00} // <- too small

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, ConnectionClose, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, DetectionError, toError(err).Code())

	<-done
}

func abstractConnDecodeError2Test(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		buf := []byte{0x10, 0xff, 0xff, 0xff, 0x80} // <- too long

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, ConnectionClose, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, DetectionError, toError(err).Code())

	<-done
}

func abstractConnDecodeError3Test(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		buf := []byte{0x20, 0x02, 0x00, 0x06} // <- invalid packet

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, ConnectionClose, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, DecodeError, toError(err).Code())

	<-done
}

func abstractConnSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	err = conn2.Send(packet.NewConnectPacket())
	assert.Equal(t, NetworkError, toError(err).Code())

	<-done
}

func abstractConnSendAndCloseTest(t *testing.T, protocol string) {
	wait := make(chan struct{})

	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Send(packet.NewConnectPacket())
		assert.NoError(t, err)

		err = conn1.Close()
		assert.NoError(t, err)

		close(wait)
	})

	<-wait

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNECT)
	assert.NoError(t, err)

	pkt, err = conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnCountersTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.NoError(t, err)
		assert.Equal(t, int64(pkt.Len()), conn1.BytesRead())

		pkt2 := packet.NewConnackPacket()
		conn1.Send(pkt2)
		assert.Equal(t, int64(pkt2.Len()), conn1.BytesWritten())

		err = conn1.Close()
		assert.NoError(t, err)
	})

	pkt := packet.NewConnectPacket()
	conn2.Send(pkt)
	assert.Equal(t, int64(pkt.Len()), conn2.BytesWritten())

	pkt2, err := conn2.Receive()
	assert.NoError(t, err)
	assert.Equal(t, int64(pkt2.Len()), conn2.BytesRead())

	err = conn2.Close()
	assert.NoError(t, err)

	<-done
}

func abstractConnReadLimitTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		conn1.SetReadLimit(1)

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, NetworkError, toError(err).Code())
		assert.Equal(t, ErrReadLimitExceeded, toError(err).Err())
	})

	err := conn2.Send(packet.NewConnectPacket())
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnReadTimeoutTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		conn1.SetReadTimeout(10 * time.Millisecond)

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, NetworkError, toError(err).Code())
		assert.Equal(t, ErrReadTimeout, toError(err).Err())
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnCloseAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)

		err = conn1.Close()
		assert.Equal(t, NetworkError, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnAddrTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		assert.NotEmpty(t, conn1.LocalAddr().String())
		assert.NotEmpty(t, conn1.RemoteAddr().String())

		err := conn1.Close()
		assert.NoError(t, err)
	})

	assert.NotEmpty(t, conn2.LocalAddr().String())
	assert.NotEmpty(t, conn2.RemoteAddr().String())

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	<-done
}

func abstractConnBufferedSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.BufferedSend(packet.NewConnackPacket())
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, ConnectionClose, toError(err).Code())
	})

	err := conn2.BufferedSend(packet.NewConnectPacket())
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	<-done
}

func abstractConnBufferedSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	err = conn2.BufferedSend(packet.NewConnectPacket())
	assert.NoError(t, err)

	<-time.After(2 * flushTimeout)

	err = conn2.BufferedSend(packet.NewConnectPacket())
	assert.Equal(t, NetworkError, toError(err).Code())

	<-done
}

func abstractConnBigBufferedSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, ConnectionClose, toError(err).Code())

	pub := packet.NewPublishPacket()
	pub.Message.Topic = "hello"
	pub.Message.Payload = make([]byte, 6400) // <- bigger than write buffer

	err = conn2.BufferedSend(pub)
	assert.Equal(t, NetworkError, toError(err).Code())

	<-done
}
