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

		packet = []byte{
			byte(SUBACK << 4),
			1, // < wrong remaining length
			0, // packet id
			7,
			0, // return code 1
		}

		pkt = NewSuback()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(SUBACK << 4),
			6, // remaining length
			0, // packet id
			7,
			0,    // return code 1
			1,    // return code 2
			2,    // return code 3
			0x81, // < wrong return code
		}

		pkt = NewSuback()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(SUBACK << 4),
			1, // < wrong remaining length
			0, // packet id
		}

		pkt = NewSuback()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(PUBCOMP << 4), // < wrong packet type
			3,                  // remaining length
			0,                  // packet id
			7,
			0, // return code 1
		}

		pkt = NewSuback()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(SUBACK << 4),
			3, // remaining length
			0, // packet id
			0, // < zero packet id
			0,
		}

		pkt = NewSuback()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)
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

		pkt = NewSuback()
		pkt.ID = 7
		pkt.ReturnCodes = []QOS{0x81}

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 0, n)

		pkt = NewSuback()
		pkt.ID = 7
		pkt.ReturnCodes = []QOS{0x80}

		dst = make([]byte, pkt.Len(m)-1)
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 0, n)

		pkt = NewSuback()
		pkt.ID = 0 // < zero packet id
		pkt.ReturnCodes = []QOS{0x80}

		dst = make([]byte, pkt.Len(m)-1)
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 0, n)
	})
}

func BenchmarkSubackEncode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		pkt := NewSuback()
		pkt.ID = 1
		pkt.ReturnCodes = []QOS{0}

		buf := make([]byte, pkt.Len(m))

		for i := 0; i < b.N; i++ {
			_, err := pkt.Encode(m, buf)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkSubackDecode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		packet := []byte{
			byte(SUBACK << 4),
			3, // remaining length
			0, // packet id
			1,
			0, // return code 1
		}

		pkt := NewSuback()

		for i := 0; i < b.N; i++ {
			_, err := pkt.Decode(m, packet)
			if err != nil {
				panic(err)
			}
		}
	})
}
