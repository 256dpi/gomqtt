package broker

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/spec"
)

func TestBrokerWithMemoryBackend(t *testing.T) {
	backend := NewMemoryBackend()
	backend.Credentials = map[string]string{
		"allow": "allow",
	}

	port, quit, done := Run(NewEngineWithBackend(backend), "tcp")

	config := spec.AllFeatures()
	config.URL = "tcp://allow:allow@localhost:" + port
	config.DenyURL = "tcp://deny:deny@localhost:" + port
	config.ProcessWait = 20 * time.Millisecond
	config.NoMessageWait = 50 * time.Millisecond
	config.MessageRetainWait = 50 * time.Millisecond

	spec.Run(t, config)

	close(quit)

	safeReceive(done)
}
