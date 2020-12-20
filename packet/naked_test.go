package packet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testNaked(t *testing.T, typ Type) {
	pkt, err := typ.New()
	assert.NoError(t, err)
	assert.Equal(t, typ, pkt.Type())
	assert.Equal(t, fmt.Sprintf("<%s>", pkt.Type().String()), pkt.String())

	buf := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)

	n, err = pkt.Decode(M4, buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestDisconnect(t *testing.T) {
	testNaked(t, DISCONNECT)
}

func TestPingreq(t *testing.T) {
	testNaked(t, PINGREQ)
}

func TestPingresp(t *testing.T) {
	testNaked(t, PINGRESP)
}

func TestNakedDecode(t *testing.T) {
	packet := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedDecode(packet, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestNakedDecodeError1(t *testing.T) {
	packet := []byte{
		byte(DISCONNECT << 4),
		1, // < wrong remaining length
		0,
	}

	n, err := nakedDecode(packet, DISCONNECT)

	assert.Error(t, err)
	assert.Equal(t, 2, n)
}

func TestNakedEncode(t *testing.T) {
	packet := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	dst := make([]byte, nakedLen())
	n, err := nakedEncode(dst, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, packet, dst[:n])
}

func BenchmarkNakedEncode(b *testing.B) {
	b.ReportAllocs()

	buf := make([]byte, nakedLen())

	for i := 0; i < b.N; i++ {
		_, err := nakedEncode(buf, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkNakedDecode(b *testing.B) {
	b.ReportAllocs()

	packet := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	for i := 0; i < b.N; i++ {
		_, err := nakedDecode(packet, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}
