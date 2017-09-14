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
	"io"
	"testing"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func abstractConnConnectTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
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
	assert.Equal(t, io.EOF, err)

	<-done
}

func abstractConnEncodeErrorTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt := packet.NewConnackPacket()
		pkt.ReturnCode = 11 // <- invalid return code

		err := conn1.Send(pkt)
		assert.Error(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	<-done
}

func abstractConnDecodeErrorTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		buf := []byte{0x00, 0x00} // <- too small

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func abstractConnSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	err = conn2.Send(packet.NewConnectPacket())
	assert.Error(t, err)

	<-done
}

func abstractConnCloseWhileSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Send(packet.NewConnectPacket())
		assert.NoError(t, err)

		err = conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.NotNil(t, pkt)
	assert.NoError(t, err)

	for {
		// keep writing
		err := conn2.Send(packet.NewConnectPacket())
		if err != nil {
			assert.Error(t, err)
			break
		}
	}

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
	assert.Equal(t, io.EOF, err)

	<-done
}

func abstractConnReadLimitTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		conn1.SetReadLimit(1)

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Error(t, err)
		assert.Equal(t, packet.ErrReadLimitExceeded, err)
	})

	err := conn2.Send(packet.NewConnectPacket())
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	<-done
}

func abstractConnReadTimeoutTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		conn1.SetReadTimeout(10 * time.Millisecond)

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Error(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func abstractConnCloseAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)

		err = conn1.Close()
		assert.Error(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

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
	assert.Error(t, err)

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
		assert.Equal(t, io.EOF, err)
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

func abstractConnSendAfterBufferedSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.BufferedSend(packet.NewConnackPacket())
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	err := conn2.BufferedSend(packet.NewConnectPacket())
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	pkt, err = conn2.Receive()
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
	assert.Equal(t, io.EOF, err)

	err = conn2.BufferedSend(packet.NewConnectPacket())
	assert.NoError(t, err)

	<-time.After(2 * flushTimeout)

	err = conn2.BufferedSend(packet.NewConnectPacket())
	assert.Error(t, err)

	<-done
}

func abstractConnCloseAfterBufferedSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	err := conn2.BufferedSend(packet.NewConnectPacket())
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	<-done
}

func abstractConnBigBufferedSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	pub := packet.NewPublishPacket()
	pub.Message.Topic = "hello"
	pub.Message.Payload = make([]byte, 6400) // <- bigger than write buffer

	err = conn2.BufferedSend(pub)
	assert.Error(t, err)

	<-done
}
