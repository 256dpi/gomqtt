package packet

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func binaryVarUintLen(n uint64) int {
	return binary.PutUvarint(make([]byte, binary.MaxVarintLen64), n)
}

func TestVarUintLen(t *testing.T) {
	assert.Equal(t, 1, varUintLen(1))
	assert.Equal(t, 1, varUintLen(127))
	assert.Equal(t, 2, varUintLen(128))
	assert.Equal(t, 2, varUintLen(16383))
	assert.Equal(t, 3, varUintLen(16384))
	assert.Equal(t, 3, varUintLen(2097151))
	assert.Equal(t, 4, varUintLen(2097152))
	assert.Equal(t, 4, varUintLen(268435455))
	assert.Equal(t, 0, varUintLen(268435456))
	assert.Equal(t, 0, varUintLen(math.MaxUint64))

	assert.Equal(t, 1, binaryVarUintLen(1))
	assert.Equal(t, 1, binaryVarUintLen(127))
	assert.Equal(t, 2, binaryVarUintLen(128))
	assert.Equal(t, 2, binaryVarUintLen(16383))
	assert.Equal(t, 3, binaryVarUintLen(16384))
	assert.Equal(t, 3, binaryVarUintLen(2097151))
	assert.Equal(t, 4, binaryVarUintLen(2097152))
	assert.Equal(t, 4, binaryVarUintLen(268435455))
	assert.Equal(t, 5, binaryVarUintLen(268435456))
	assert.Equal(t, 10, binaryVarUintLen(math.MaxUint64))
}
