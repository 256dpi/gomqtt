package transport

import (
	"io"
	"testing"

	"github.com/256dpi/gomqtt/packet"

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

func TestWebSocketConnDecodeError(t *testing.T) {
	abstractConnDecodeErrorTest(t, "ws")
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

func TestWebSocketConnAsyncSend(t *testing.T) {
	abstractConnAsyncSendTest(t, "ws")
}

func TestWebSocketConnSendAfterAsyncSend(t *testing.T) {
	abstractConnSendAfterAsyncSendTest(t, "ws")
}

func TestWebSocketConnAsyncSendAfterClose(t *testing.T) {
	abstractConnAsyncSendAfterCloseTest(t, "ws")
}

func TestWebSocketConnCloseAfterAsyncSend(t *testing.T) {
	abstractConnCloseAfterAsyncSendTest(t, "ws")
}

func TestWebSocketConnBigAsyncSendAfterClose(t *testing.T) {
	abstractConnBigAsyncSendAfterCloseTest(t, "ws")
}

func TestWebSocketBadFrameError(t *testing.T) {
	conn2, done := connectionPair("ws", func(conn1 Conn) {
		buf := []byte{0x07, 0x00, 0x00, 0x00, 0x00} // < bad frame

		_, err := conn1.(*WebSocketConn).UnderlyingConn().UnderlyingConn().Write(buf)
		assert.NoError(t, err)

		pkt, err := conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	safeReceive(done)
}

func TestWebSocketChunkedMessage(t *testing.T) {
	pkt := packet.NewPublish()
	pkt.Message.Topic = "hello"
	pkt.Message.Payload = []byte("world")

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		buf := make([]byte, pkt.Len())
		_, err := pkt.Encode(buf)
		assert.NoError(t, err)

		err = conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.BinaryMessage, buf[:7])
		assert.NoError(t, err)

		err = conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.BinaryMessage, buf[7:])
		assert.NoError(t, err)

		in, err := conn1.Receive()
		assert.Nil(t, in)
		assert.Equal(t, io.EOF, err)
	})

	in, err := conn2.Receive()
	assert.Nil(t, err)
	assert.Equal(t, pkt.String(), in.String())

	err = conn2.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestWebSocketCoalescedMessage(t *testing.T) {
	pkt := packet.NewPublish()
	pkt.Message.Topic = "hello"
	pkt.Message.Payload = []byte("world")

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		buf := make([]byte, pkt.Len()*2)

		_, err := pkt.Encode(buf)
		assert.NoError(t, err)

		_, err = pkt.Encode(buf[pkt.Len():])
		assert.NoError(t, err)

		err = conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.BinaryMessage, buf)
		assert.NoError(t, err)

		in, err := conn1.Receive()
		assert.Nil(t, in)
		assert.Equal(t, io.EOF, err)
	})

	in, err := conn2.Receive()
	assert.Nil(t, err)
	assert.Equal(t, pkt.String(), in.String())

	in, err = conn2.Receive()
	assert.Nil(t, err)
	assert.Equal(t, pkt.String(), in.String())

	err = conn2.Close()
	assert.NoError(t, err)

	safeReceive(done)
}

func TestWebSocketNotBinaryMessage(t *testing.T) {
	pkt := packet.NewPublish()
	pkt.Message.Topic = "hello"
	pkt.Message.Payload = []byte("world")

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		err := conn1.(*WebSocketConn).UnderlyingConn().WriteMessage(websocket.TextMessage, []byte("hello"))
		assert.NoError(t, err)
	})

	in, err := conn2.Receive()
	assert.Error(t, err)
	assert.Nil(t, in)

	safeReceive(done)
}

func BenchmarkWebSocketConn(b *testing.B) {
	pkt := packet.NewPublish()
	pkt.Message.Topic = "foo/bar/baz"

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		for i := 0; i < b.N; i++ {
			err := conn1.Send(pkt, false)
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

	safeReceive(done)
}

func BenchmarkWebSocketConnAsync(b *testing.B) {
	pkt := packet.NewPublish()
	pkt.Message.Topic = "foo/bar/baz"

	conn2, done := connectionPair("ws", func(conn1 Conn) {
		for i := 0; i < b.N; i++ {
			err := conn1.Send(pkt, true)
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

	safeReceive(done)
}
