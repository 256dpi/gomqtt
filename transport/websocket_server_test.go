package transport

import (
	"io/ioutil"
	"net/http"
	"testing"

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
	abstractServerCloseAfterCloseTest(t, "ws")
}

func TestWebSocketServerAddr(t *testing.T) {
	abstractServerAddrTest(t, "ws")
}

func TestWebSocketServerInvalidUpgrade(t *testing.T) {
	server, err := testLauncher.Launch("ws://localhost:0")
	require.NoError(t, err)

	res, err := http.Get(getURL(server, "http"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	err = server.Close()
	assert.NoError(t, err)
}

func TestWebSocketServerAcceptAfterError(t *testing.T) {
	server, err := testLauncher.Launch("ws://localhost:0")
	require.NoError(t, err)

	webSocketServer := server.(*WebSocketServer)

	err = webSocketServer.listener.Close()
	assert.NoError(t, err)

	conn, err := server.Accept()
	require.Nil(t, conn)
	assert.Error(t, err)
}

func TestWebSocketServerConnectionCancelOnClose(t *testing.T) {
	server, err := testLauncher.Launch("ws://localhost:0")
	require.NoError(t, err)

	conn, err := testDialer.Dial(getURL(server, "ws"))
	require.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)
}

func TestWebSocketFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello world: " + r.Proto))
	})

	server, err := CreateWebSocketServer("localhost:0", mux)
	require.NoError(t, err)

	resp, err := http.Get(getURL(server, "http") + "/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello world: HTTP/1.1"), bytes)

	err = server.Close()
	assert.NoError(t, err)
}

func TestWebSocketFallbackHTTP2(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello world: " + r.Proto))
	})

	server, err := CreateSecureWebSocketServer("localhost:0", testTLSConfig, mux)
	require.NoError(t, err)

	resp, err := http.Get(getURL(server, "https") + "/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello world: HTTP/2.0"), bytes)

	err = server.Close()
	assert.NoError(t, err)
}
