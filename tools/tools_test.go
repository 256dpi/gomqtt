package tools

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	counter := NewCounter()

	assert.Equal(t, uint16(1), counter.Next())
	assert.Equal(t, uint16(2), counter.Next())

	for i := 0; i < math.MaxUint16-3; i++ {
		counter.Next()
	}

	assert.Equal(t, uint16(math.MaxUint16), counter.Next())
	assert.Equal(t, uint16(1), counter.Next())

	counter.Reset()

	assert.Equal(t, uint16(1), counter.Next())
}
