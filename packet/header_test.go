package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderDecodeError1(t *testing.T) {
	buf := []byte{0x6f, 193, 2} // < not enough bytes

	_, _, _, err := headerDecode(buf, 0)
	assert.Error(t, err)
}

func TestHeaderDecodeError2(t *testing.T) {
	// source to small
	buf := []byte{0x62}

	_, _, _, err := headerDecode(buf, 0)
	assert.Error(t, err)
}

func TestHeaderDecodeError3(t *testing.T) {
	buf := []byte{0x62, 0xff} // < invalid packet type

	_, _, _, err := headerDecode(buf, 0)
	assert.Error(t, err)
}

func TestHeaderDecodeError4(t *testing.T) {
	// remaining length to big
	buf := []byte{0x62, 0xff, 0xff, 0xff, 0xff}

	n, _, _, err := headerDecode(buf, 6)

	assert.Error(t, err)
	assert.Equal(t, 1, n)
}

func TestHeaderDecodeError5(t *testing.T) {
	buf := []byte{0x66, 0x00, 0x01} // < wrong flags

	n, _, _, err := headerDecode(buf, 6)
	assert.Error(t, err)
	assert.Equal(t, 1, n)
}

func TestHeaderEncode1(t *testing.T) {
	headerBytes := []byte{0x62, 193, 2}

	buf := make([]byte, 3)
	n, err := headerEncode(buf, 0, 321, 3, PUBREL)

	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, headerBytes, buf)
}

func TestHeaderEncode2(t *testing.T) {
	headerBytes := []byte{0x62, 0xff, 0xff, 0xff, 0x7f}

	buf := make([]byte, 5)
	n, err := headerEncode(buf, 0, maxRemainingLength, 5, PUBREL)

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, headerBytes, buf)
}

func TestHeaderEncodeError1(t *testing.T) {
	headerBytes := []byte{0x00}

	buf := make([]byte, 1) // < wrong buffer size
	n, err := headerEncode(buf, 0, 0, 2, PUBREL)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, headerBytes, buf)
}

func TestHeaderEncodeError2(t *testing.T) {
	headerBytes := []byte{0x00, 0x00}

	buf := make([]byte, 2)
	// overload max remaining length
	n, err := headerEncode(buf, 0, maxRemainingLength+1, 2, PUBREL)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, headerBytes, buf)
}

func TestHeaderEncodeError3(t *testing.T) {
	headerBytes := []byte{0x00, 0x00}

	buf := make([]byte, 2)
	// too small buffer
	n, err := headerEncode(buf, 0, 2097151, 2, PUBREL)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, headerBytes, buf)
}
