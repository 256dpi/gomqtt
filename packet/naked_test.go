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
		assert.Equal(t, pkt.Len(m), n)

		n, err = pkt.Decode(m, buf)
		assert.NoError(t, err)
		assert.Equal(t, pkt.Len(m), n)
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
			0, // remaining length
		}

		n, err := nakedDecode(m, packet, DISCONNECT)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)

		assertDecodeError(t, m, DISCONNECT, 2, []byte{
			byte(DISCONNECT << 4),
			1, // < wrong remaining length
			0,
		})

		assertDecodeError(t, m, DISCONNECT, 2, []byte{
			byte(DISCONNECT << 4),
			1, // remaining length
			0, // < superfluous byte
		})
	})
}

func TestNakedEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(DISCONNECT << 4),
			0, // remaining length
		}

		dst := make([]byte, nakedLen())
		n, err := nakedEncode(m, dst, DISCONNECT)
		assert.NoError(t, err)
		assert.Equal(t, nakedLen(), n)
		assert.Equal(t, packet, dst)

		// small buffer
		assertEncodeError(t, m, 1, 1, &Disconnect{})
	})
}

func BenchmarkNaked(b *testing.B) {
	benchPacket(b, &Disconnect{})
}
