package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {
	buf := []byte{0x62, 0x1, 0x0}
	n, flags, rl, err := decodeHeader(buf, PUBREL)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, uint8(0x2), flags)
	assert.Equal(t, 1, rl)

	buf = []byte{}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = []byte{0x0}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, ErrInvalidPacketType, err)

	buf = []byte{0x6f}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, ErrInvalidStaticFlags, err)

	buf = []byte{0x62}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = []byte{0x62, 0xff}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = []byte{0x62, 0xff, 0xff, 0xff, 0xff}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = []byte{0x62, 0xff, 0xff, 0xff, 0xff, 0x1}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, ErrVariableIntegerOverflow, err)

	buf = []byte{0x62, 0x0, 0x0}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, ErrRemainingLengthMismatch, err)

	buf = []byte{0x62, 0xff, 0x1}
	n, _, _, err = decodeHeader(buf, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, ErrRemainingLengthMismatch, err)
}

func TestEncodeHeader(t *testing.T) {
	buf := make([]byte, 3+321)
	n, err := encodeHeader(buf, 0, 321, PUBREL)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{0x62, 193, 2}, buf[:3])

	buf = make([]byte, 5+maxVarint)
	n, err = encodeHeader(buf, 0, maxVarint, PUBREL)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{0x62, 0xff, 0xff, 0xff, 0x7f}, buf[:5])

	buf = make([]byte, 0)
	n, err = encodeHeader(buf, 0, 0, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, []byte{}, buf)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = make([]byte, 0)
	n, err = encodeHeader(buf, 0, 0, 0x80)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, []byte{}, buf)
	assert.Equal(t, ErrInvalidPacketType, err)

	buf = make([]byte, 0)
	n, err = encodeHeader(buf, 1, 0, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, []byte{}, buf)
	assert.Equal(t, ErrInvalidStaticFlags, err)

	buf = make([]byte, 1)
	n, err = encodeHeader(buf, 0, 0, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte{0x62}, buf)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = make([]byte, 2)
	n, err = encodeHeader(buf, 0, 2097151, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte{0x62, 0x00}, buf)
	assert.Equal(t, ErrInsufficientBufferSize, err)

	buf = make([]byte, 2)
	n, err = encodeHeader(buf, 0, maxVarint+1, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte{0x62, 0x00}, buf)
	assert.Equal(t, ErrVariableIntegerOverflow, err)

	buf = make([]byte, 3)
	n, err = encodeHeader(buf, 0, 0, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, ErrRemainingLengthMismatch, err)
}
