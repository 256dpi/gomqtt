package transport

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalDial(t *testing.T) {
	server, err := testLauncher.Launch("tcp://localhost:0")
	require.NoError(t, err)

	wait := make(chan struct{})

	go func() {
		conn, err := server.Accept()
		require.NoError(t, err)

		pkt, err := conn.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)

		close(wait)
	}()

	conn, err := Dial(getURL(server, "tcp"))
	require.NoError(t, err)

	err = conn.Close()
	assert.NoError(t, err)

	<-wait

	err = server.Close()
	assert.NoError(t, err)
}

func TestDialerBadURL(t *testing.T) {
	conn, err := Dial("foo")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestDialerUnsupportedProtocol(t *testing.T) {
	conn, err := Dial("foo://localhost")
	assert.Nil(t, conn)
	assert.Equal(t, ErrUnsupportedProtocol, err)
}

func TestDialerTCPError(t *testing.T) {
	conn, err := Dial("tcp://localhost:1234567")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestDialerTLSError(t *testing.T) {
	conn, err := Dial("tls://localhost:1234567")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestDialerWSError(t *testing.T) {
	conn, err := Dial("ws://localhost:1234567")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestDialerWSSError(t *testing.T) {
	conn, err := Dial("wss://localhost:1234567")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func abstractDefaultPortTest(t *testing.T, protocol string) {
	server, err := testLauncher.Launch(protocol + "://localhost:0")
	require.NoError(t, err)

	go func() {
		conn, err := server.Accept()
		require.NoError(t, err)

		pkt, err := conn.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	}()

	dialer := NewDialer()
	dialer.TLSConfig = clientTLSConfig
	dialer.DefaultTCPPort = getPort(server)
	dialer.DefaultTLSPort = getPort(server)
	dialer.DefaultWSPort = getPort(server)
	dialer.DefaultWSSPort = getPort(server)

	conn, err := dialer.Dial(protocol + "://localhost")
	require.NoError(t, err)

	err = conn.Close()
	assert.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func TestTCPDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "tcp")
}

func TestTLSDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "tls")
}

func TestWSDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "ws")
}

func TestWSSDefaultPort(t *testing.T) {
	abstractDefaultPortTest(t, "wss")
}
