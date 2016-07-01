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
	config.MessageRetainWait = 300 * time.Millisecond
	config.NoMessageWait = 100 * time.Millisecond

	Run(t, config)
}
