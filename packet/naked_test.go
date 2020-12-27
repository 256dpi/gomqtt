package packet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testNaked(t *testing.T, typ Type) {
	multiTest(t, func(t *testing.T, m Mode) {
		pkt, err := typ.New()
		assert.NoError(t, err)
		assert.Equal(t, typ, pkt.Type())
		assert.Equal(t, fmt.Sprintf("<%s>", pkt.Type().String()), pkt.String())

		buf := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, buf)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)

		n, err = pkt.Decode(m, buf)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
	})
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
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(DISCONNECT << 4),
			0,
		}

		n, err := nakedDecode(m, packet, DISCONNECT)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
	})
}

func TestNakedDecodeError1(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(DISCONNECT << 4),
			1, // < wrong remaining length
			0,
		}

		n, err := nakedDecode(m, packet, DISCONNECT)

		assert.Error(t, err)
		assert.Equal(t, 2, n)
	})
}

func TestNakedEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(DISCONNECT << 4),
			0,
		}

		dst := make([]byte, nakedLen())
		n, err := nakedEncode(m, dst, DISCONNECT)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.Equal(t, packet, dst[:n])
	})
}

func BenchmarkNakedEncode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		buf := make([]byte, nakedLen())

		for i := 0; i < b.N; i++ {
			_, err := nakedEncode(m, buf, DISCONNECT)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkNakedDecode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		packet := []byte{
			byte(DISCONNECT << 4),
			0,
		}

		for i := 0; i < b.N; i++ {
			_, err := nakedDecode(m, packet, DISCONNECT)
			if err != nil {
				panic(err)
			}
		}
	})
}
