package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuback(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		pkt := NewSuback()
		pkt.ID = 1
		pkt.ReturnCodes = []QOS{0, 1}

		assert.Equal(t, pkt.Type(), SUBACK)
		assert.Equal(t, "<Suback ID=1 ReturnCodes=[0, 1]>", pkt.String())

		buf := make([]byte, pkt.Len(m))
		n1, err := pkt.Encode(m, buf)
		assert.NoError(t, err)

		pkt2 := NewSuback()
		n2, err := pkt2.Decode(m, buf)
		assert.NoError(t, err)

		assert.Equal(t, pkt, pkt2)
		assert.Equal(t, n1, n2)
	})
}

func TestSubackDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(SUBACK << 4),
			6, // remaining length
			0, // packet id
			7,
			0,    // return code 1
			1,    // return code 2
			2,    // return code 3
			0x80, // return code 4
		}

		pkt := NewSuback()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, 4, len(pkt.ReturnCodes))

		assertDecodeError(t, m, SUBACK, 4, []byte{
			byte(SUBACK << 4),
			1, // < wrong remaining length
			0, // packet id
			7,
			0, // return code 1
		})

		assertDecodeError(t, m, SUBACK, 8, []byte{
			byte(SUBACK << 4),
			6, // remaining length
			0, // packet id
			7,
			0,    // return code 1
			1,    // return code 2
			2,    // return code 3
			0x81, // < wrong return code
		})

		assertDecodeError(t, m, SUBACK, 2, []byte{
			byte(SUBACK << 4),
			1, // < wrong remaining length
			0, // packet id
		})

		assertDecodeError(t, m, SUBACK, 1, []byte{
			byte(PUBCOMP << 4), // < wrong packet type
			3,                  // remaining length
			0,                  // packet id
			7,
			0, // return code 1
		})

		assertDecodeError(t, m, SUBACK, 4, []byte{
			byte(SUBACK << 4),
			3, // remaining length
			0, // packet id
			0, // < zero packet id
			0,
		})
	})
}

func TestSubackEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(SUBACK << 4),
			6, // remaining length
			0, // packet id
			7,
			0,    // return code 1
			1,    // return code 2
			2,    // return code 3
			0x80, // return code 4
		}

		pkt := NewSuback()
		pkt.ID = 7
		pkt.ReturnCodes = []QOS{0, 1, 2, 0x80}

		dst := make([]byte, 10)
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		// small buffer
		assertEncodeError(t, m, 1, 1, &Suback{
			ID:          7,
			ReturnCodes: []QOS{0x80},
		})

		assertEncodeError(t, m, 0, 4, &Suback{
			ID:          7,
			ReturnCodes: []QOS{0x81}, // < invalid
		})

		assertEncodeError(t, m, 0, 2, &Suback{
			ID:          0, // < zero packet id
			ReturnCodes: []QOS{0},
		})
	})
}

func BenchmarkSuback(b *testing.B) {
	benchPacket(b, &Suback{
		ReturnCodes: []QOS{0},
		ID:          7,
	})
}
