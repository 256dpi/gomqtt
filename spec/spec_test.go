package spec

import (
	"testing"
	"time"
)

func TestSpec(t *testing.T) {
	config := AllFeatures()
	config.URL = "tcp://localhost:1883"

	// mosquitto specific config
	config.Authentication = false
	config.ProcessWait = 10 * time.Millisecond
	config.MessageRetainWait = 100 * time.Millisecond
	config.NoMessageWait = 50 * time.Millisecond

	Run(t, config)
}
