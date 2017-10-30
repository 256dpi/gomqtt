package tools

import (
	"math"
	"testing"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	counter := NewCounter()

	assert.Equal(t, packet.ID(1), counter.Next())
	assert.Equal(t, packet.ID(2), counter.Next())

	for i := 0; i < math.MaxUint16-3; i++ {
		counter.Next()
	}

	assert.Equal(t, packet.ID(math.MaxUint16), counter.Next())
	assert.Equal(t, packet.ID(1), counter.Next())

	counter.Reset()

	assert.Equal(t, packet.ID(1), counter.Next())
}
