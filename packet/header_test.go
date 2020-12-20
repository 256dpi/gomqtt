package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeaderError1(t *testing.T) {
	buf := []byte{0x6f, 193, 2} // < not enough bytes

	_, _, _, err := decodeHeader(buf, 0)
	assert.Error(t, err)
}

func TestDecodeHeaderError2(t *testing.T) {
	// source to small
	buf := []byte{0x62}

	_, _, _, err := decodeHeader(buf, 0)
	assert.Error(t, err)
}

func TestDecodeHeaderError3(t *testing.T) {
	buf := []byte{0x62, 0xff} // < invalid packet type

	_, _, _, err := decodeHeader(buf, 0)
	assert.Error(t, err)
}

func TestDecodeHeaderError4(t *testing.T) {
	// invalid remaining length
	buf := []byte{0x62, 0xff, 0xff, 0xff, 0xff}

	n, _, _, err := decodeHeader(buf, 6)

	assert.Error(t, err)
	assert.Equal(t, 1, n)
}

func TestDecodeHeaderError5(t *testing.T) {
	// remaining length to big
	buf := []byte{0x62, 0xff, 0xff, 0xff, 0xff, 0x1}

	n, _, _, err := decodeHeader(buf, 6)

	assert.Error(t, err)
	assert.Equal(t, 1, n)
}

func TestDecodeHeaderError6(t *testing.T) {
	// source to small
	buf := []byte{0x62, 0xff, 0x1}

	n, _, _, err := decodeHeader(buf, 6)

	assert.Error(t, err)
	assert.Equal(t, 3, n)
}

func TestDecodeHeaderError7(t *testing.T) {
	buf := []byte{0x66, 0x00, 0x01} // < wrong flags

	n, _, _, err := decodeHeader(buf, 6)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
}

func TestEncodeHeader1(t *testing.T) {
	buf := make([]byte, 3)
	n, err := encodeHeader(buf, 0, 321, 3, PUBREL)

	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{0x62, 193, 2}, buf)
}

func TestEncodeHeader2(t *testing.T) {
	buf := make([]byte, 5)
	n, err := encodeHeader(buf, 0, maxVarint, 5, PUBREL)

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{0x62, 0xff, 0xff, 0xff, 0x7f}, buf)
}

func TestEncodeHeaderError1(t *testing.T) {
	buf := make([]byte, 1) // < wrong buffer size
	n, err := encodeHeader(buf, 0, 0, 2, PUBREL)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, []byte{0x00}, buf)
}

func TestEncodeHeaderError2(t *testing.T) {
	buf := make([]byte, 2)
	// overload max remaining length
	n, err := encodeHeader(buf, 0, maxVarint+1, 2, PUBREL)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, []byte{0x62, 0x00}, buf)
}

func TestEncodeHeaderError3(t *testing.T) {
	buf := make([]byte, 2)
	// too small buffer
	n, err := encodeHeader(buf, 0, 2097151, 2, PUBREL)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, []byte{0x00, 0x00}, buf)
}
