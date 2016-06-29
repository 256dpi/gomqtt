package spec

import "testing"

func TestSpec(t *testing.T) {
	matrix := FullMatrix

	// mosquitto does not support authentication
	matrix.Authentication = false

	Run(t, matrix, "localhost:1883")
}
