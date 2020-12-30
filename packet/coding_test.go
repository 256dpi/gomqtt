package packet

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarintLen(t *testing.T) {
	nativeVarintLen := func(n uint64) int {
		return binary.PutUvarint(make([]byte, binary.MaxVarintLen64), n)
	}

	assert.Equal(t, 1, varintLen(1))
	assert.Equal(t, 1, varintLen(127))
	assert.Equal(t, 2, varintLen(128))
	assert.Equal(t, 2, varintLen(16383))
	assert.Equal(t, 3, varintLen(16384))
	assert.Equal(t, 3, varintLen(2097151))
	assert.Equal(t, 4, varintLen(2097152))
	assert.Equal(t, 4, varintLen(268435455))
	assert.Equal(t, 0, varintLen(268435456))
	assert.Equal(t, 0, varintLen(math.MaxUint64))

	assert.Equal(t, 1, nativeVarintLen(1))
	assert.Equal(t, 1, nativeVarintLen(127))
	assert.Equal(t, 2, nativeVarintLen(128))
	assert.Equal(t, 2, nativeVarintLen(16383))
	assert.Equal(t, 3, nativeVarintLen(16384))
	assert.Equal(t, 3, nativeVarintLen(2097151))
	assert.Equal(t, 4, nativeVarintLen(2097152))
	assert.Equal(t, 4, nativeVarintLen(268435455))
	assert.Equal(t, 5, nativeVarintLen(268435456))
	assert.Equal(t, 10, nativeVarintLen(math.MaxUint64))
}

func TestReadVarint(t *testing.T) {
	num, n, err := readVarint([]byte{0xff, 0x42})
	assert.NoError(t, err)
	assert.Equal(t, 8575, int(num))
	assert.Equal(t, 2, n)

	num, n, err = readVarint([]byte{})
	assert.Error(t, err)
	assert.Zero(t, num)
	assert.Zero(t, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	num, n, err = readVarint([]byte{0xff, 0xff})
	assert.Error(t, err)
	assert.Zero(t, num)
	assert.Equal(t, 0, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	num, n, err = readVarint([]byte{0xff, 0xff, 0xff, 0xff, 0x1})
	assert.Error(t, err)
	assert.Zero(t, num)
	assert.Equal(t, 0, n)
	assert.Equal(t, ErrVariableIntegerOverflow, err)
}

func TestWriteVarint(t *testing.T) {
	n, err := writeVarint([]byte{0, 0}, 8575)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)

	n, err = writeVarint([]byte{}, 42)
	assert.Error(t, err)
	assert.Zero(t, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	n, err = writeVarint([]byte{0}, 8575)
	assert.Error(t, err)
	assert.Zero(t, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	n, err = writeVarint([]byte{0, 0, 0, 0}, maxVarint+1)
	assert.Error(t, err)
	assert.Zero(t, n)
	assert.Equal(t, ErrVariableIntegerOverflow, err)
}

func TestReadString(t *testing.T) {
	str, n, err := readString([]byte{0x0, 0x0})
	assert.NoError(t, err)
	assert.Equal(t, "", str)
	assert.Equal(t, 2, n)

	str, n, err = readString([]byte{0x0, 0x5, 'H', 'e', 'l', 'l', 'o'})
	assert.NoError(t, err)
	assert.Equal(t, "Hello", str)
	assert.Equal(t, 7, n)

	str, n, err = readString([]byte{})
	assert.Error(t, err)
	assert.Empty(t, str)
	assert.Zero(t, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	str, n, err = readString([]byte{0xff, 0xff, 0xff, 0xff})
	assert.Error(t, err)
	assert.Empty(t, str)
	assert.Equal(t, 2, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)
}

func TestWriteString(t *testing.T) {
	buf := make([]byte, 7)
	n, err := writeString(buf, "Hello")
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x0, 0x5, 'H', 'e', 'l', 'l', 'o'}, buf)
	assert.Equal(t, 7, n)

	n, err = writeString([]byte{}, longString)
	assert.Error(t, err)
	assert.Zero(t, n)
	assert.Equal(t, ErrPrefixedBytesOverflow, err)

	n, err = writeString([]byte{}, string(make([]byte, 10)))
	assert.Error(t, err)
	assert.Zero(t, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)
}
