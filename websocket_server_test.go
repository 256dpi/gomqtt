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
