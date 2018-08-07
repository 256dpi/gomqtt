package transport

import (
	"io"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func abstractConnConnectTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnack(), false)
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	err := conn2.Send(packet.NewConnect(), false)
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func abstractConnCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	safeReceive(done)
}

func abstractConnEncodeErrorTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt := packet.NewConnack()
		pkt.ReturnCode = 11 // < invalid return code

		err := conn1.Send(pkt, false)
		assert.Error(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	safeReceive(done)
}

func abstractConnDecodeErrorTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		buf := []byte{0x00, 0x00} // < too small

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

	safeReceive(done)
}

func abstractConnSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	err = conn2.Send(packet.NewConnect(), false)
	assert.Error(t, err)

	safeReceive(done)
}

func abstractConnCloseWhileSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Send(packet.NewConnect(), false)
		assert.NoError(t, err)

		err = conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.NotNil(t, pkt)
	assert.NoError(t, err)

	for {
		// keep writing
		err := conn2.Send(packet.NewConnect(), false)
		if err != nil {
			assert.Error(t, err)
			break
		}
	}

	safeReceive(done)
}

func abstractConnSendAndCloseTest(t *testing.T, protocol string) {
	wait := make(chan struct{})

	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Send(packet.NewConnect(), false)
		assert.NoError(t, err)

		err = conn1.Close()
		assert.NoError(t, err)

		close(wait)
	})

	safeReceive(wait)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNECT)
	assert.NoError(t, err)

	pkt, err = conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	safeReceive(done)
}

func abstractConnReadLimitTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		conn1.SetReadLimit(1)

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Error(t, err)
		assert.Equal(t, packet.ErrReadLimitExceeded, err)
	})

	err := conn2.Send(packet.NewConnect(), false)
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	safeReceive(done)
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

	safeReceive(done)
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

	safeReceive(done)
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

	safeReceive(done)
}

func abstractConnAsyncSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnack(), true)
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	err := conn2.Send(packet.NewConnect(), true)
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func abstractConnSendAfterAsyncSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnack(), true)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnack(), false)
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	err := conn2.Send(packet.NewConnect(), true)
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	pkt, err = conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func abstractConnAsyncSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	err = conn2.Send(packet.NewConnect(), true)
	assert.NoError(t, err)

	<-time.After(2 * FlushTimeout)

	err = conn2.Send(packet.NewConnect(), true)
	assert.Error(t, err)

	safeReceive(done)
}

func abstractConnCloseAfterAsyncSendTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	err := conn2.Send(packet.NewConnect(), true)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func abstractConnBigAsyncSendAfterCloseTest(t *testing.T, protocol string) {
	conn2, done := connectionPair(protocol, func(conn1 Conn) {
		err := conn1.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, io.EOF, err)

	pub := packet.NewPublish()
	pub.Message.Topic = "hello"
	pub.Message.Payload = make([]byte, 6400) // < bigger than write buffer

	err = conn2.Send(pub, true)
	assert.Error(t, err)

	safeReceive(done)
}
