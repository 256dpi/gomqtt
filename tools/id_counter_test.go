package tools

import (
	"math"
	"testing"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestIDCounter(t *testing.T) {
	counter := NewIDCounter()

	assert.Equal(t, packet.ID(1), counter.NextID())
	assert.Equal(t, packet.ID(2), counter.NextID())

	for i := 0; i < math.MaxUint16-3; i++ {
		counter.NextID()
	}

	assert.Equal(t, packet.ID(math.MaxUint16), counter.NextID())
	assert.Equal(t, packet.ID(1), counter.NextID())

	counter.Reset()

	assert.Equal(t, packet.ID(1), counter.NextID())
}
