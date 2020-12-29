package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {
	buf := []byte{0x6f, 193, 2} // < not enough bytes
	_, _, _, err := decodeHeader(buf, 0)
	assert.Error(t, err)

	// source to small
	buf = []byte{0x62}
	_, _, _, err = decodeHeader(buf, 0)
	assert.Error(t, err)

	buf = []byte{0x62, 0xff} // < invalid packet type
	_, _, _, err = decodeHeader(buf, 0)
	assert.Error(t, err)

	// invalid remaining length
	buf = []byte{0x62, 0xff, 0xff, 0xff, 0xff}
	n, _, _, err := decodeHeader(buf, 6)
	assert.Error(t, err)
	assert.Equal(t, 1, n)

	// remaining length to big
	buf = []byte{0x62, 0xff, 0xff, 0xff, 0xff, 0x1}
	n, _, _, err = decodeHeader(buf, 6)
	assert.Error(t, err)
	assert.Equal(t, 1, n)

	// source to small
	buf = []byte{0x62, 0xff, 0x1}
	n, _, _, err = decodeHeader(buf, 6)
	assert.Error(t, err)
	assert.Equal(t, 3, n)

	buf = []byte{0x66, 0x00, 0x01} // < wrong flags
	n, _, _, err = decodeHeader(buf, 6)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
}

func TestEncodeHeader(t *testing.T) {
	buf := make([]byte, 3)
	n, err := encodeHeader(buf, 0, 321, PUBREL)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{0x62, 193, 2}, buf)

	buf = make([]byte, 5)
	n, err = encodeHeader(buf, 0, maxVarint, PUBREL)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{0x62, 0xff, 0xff, 0xff, 0x7f}, buf)

	buf = make([]byte, 1) // < wrong buffer size
	n, err = encodeHeader(buf, 0, 0, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte{0x62}, buf)

	// overload max remaining length
	buf = make([]byte, 2)
	n, err = encodeHeader(buf, 0, maxVarint+1, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte{0x62, 0x00}, buf)

	// too small buffer
	buf = make([]byte, 2)
	n, err = encodeHeader(buf, 0, 2097151, PUBREL)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte{0x62, 0x00}, buf)
}
