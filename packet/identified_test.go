package packet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testIdentified(t *testing.T, pkt Generic) {
	multiTest(t, func(t *testing.T, m Mode) {
		assert.Equal(t, fmt.Sprintf("<%s ID=1>", pkt.Type().String()), pkt.String())

		buf := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, buf)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)

		n, err = pkt.Decode(m, buf)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
	})
}

func TestPuback(t *testing.T) {
	pkt := NewPuback()
	pkt.ID = 1

	testIdentified(t, pkt)
}

func TestPubcomp(t *testing.T) {
	pkt := NewPubcomp()
	pkt.ID = 1

	testIdentified(t, pkt)
}

func TestPubrec(t *testing.T) {
	pkt := NewPubrec()
	pkt.ID = 1

	testIdentified(t, pkt)
}

func TestPubrel(t *testing.T) {
	pkt := NewPubrel()
	pkt.ID = 1

	testIdentified(t, pkt)
}

func TestUnsuback(t *testing.T) {
	pkt := NewUnsuback()
	pkt.ID = 1

	testIdentified(t, pkt)
}

func TestIdentifiedDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(PUBACK << 4),
			2, // remaining length
			0, // packet id
			7,
		}

		var pid ID
		n, err := identifiedDecode(m, packet, &pid, PUBACK)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, ID(7), pid)

		assertDecodeError(t, m, PUBACK, 2, []byte{
			byte(PUBACK << 4),
			1, // < wrong remaining length
			0, // packet id
			7,
		})

		assertDecodeError(t, m, PUBACK, 2, []byte{
			byte(PUBACK << 4),
			2, // remaining length
			7, // packet id
			// < insufficient bytes
		})

		assertDecodeError(t, m, PUBACK, 4, []byte{
			byte(PUBACK << 4),
			2, // remaining length
			0, // packet id
			0, // < zero id
		})
	})
}

func TestIdentifiedEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(PUBACK << 4),
			2, // remaining length
			0, // packet id
			7,
		}

		dst := make([]byte, identifiedLen())
		n, err := identifiedEncode(m, dst, 7, PUBACK)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, packet, dst[:n])

		// small buffer
		assertEncodeError(t, m, 1, 0, &Puback{
			ID: 7,
		})

		assertEncodeError(t, m, 0, 2, &Puback{
			ID: 0, // < zero id
		})
	})
}

func BenchmarkIdentifiedEncode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		pkt := &Puback{}
		pkt.ID = 1

		buf := make([]byte, pkt.Len(m))

		for i := 0; i < b.N; i++ {
			_, err := pkt.Encode(m, buf)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkIdentifiedDecode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		packet := []byte{
			byte(PUBACK << 4),
			2, // remaining length
			0, // packet id
			1,
		}

		pkt := &Puback{}

		for i := 0; i < b.N; i++ {
			_, err := pkt.Decode(m, packet)
			if err != nil {
				panic(err)
			}
		}
	})
}
