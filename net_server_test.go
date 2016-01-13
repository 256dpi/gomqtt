package transport

import "testing"

func TestTCPServer(t *testing.T) {
	abstractServerTest(t, "tcp")
}

func TestTLSServer(t *testing.T) {
	abstractServerTest(t, "tls")
}

func TestTCPServerLaunchError(t *testing.T) {
	abstractServerLaunchErrorTest(t, "tcp")
}

func TestTLSServerLaunchError(t *testing.T) {
	abstractServerLaunchErrorTest(t, "tls")
}
