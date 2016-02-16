package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestPort(t *testing.T) {
	port := NewPort()
	assert.Equal(t, fmt.Sprintf("tcp://localhost:%s/", port.Port()), port.URL())
	assert.Equal(t, fmt.Sprintf("ws://localhost:%s/", port.Port()), port.URL("ws"))
}
