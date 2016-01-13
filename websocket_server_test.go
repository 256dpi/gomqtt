package transport

import "testing"

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

