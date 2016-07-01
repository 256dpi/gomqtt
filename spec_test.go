package spec

import "testing"

func TestSpec(t *testing.T) {
	config := AllFeatures()
	config.URL = "tcp://localhost:1883"

	// mosquitto does not support authentication
	config.Authentication = false

	Run(t, config)
}
