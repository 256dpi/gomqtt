package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsubscribe(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		pkt := NewUnsubscribe()
		pkt.ID = 1
		pkt.Topics = []string{"foo", "bar"}

		assert.Equal(t, pkt.Type(), UNSUBSCRIBE)
		assert.Equal(t, `<Unsubscribe ID=1 Topics=["foo", "bar"]>`, pkt.String())

		buf := make([]byte, pkt.Len(m))
		n1, err := pkt.Encode(m, buf)
		assert.NoError(t, err)

		pkt2 := NewUnsubscribe()
		n2, err := pkt2.Decode(m, buf)
		assert.NoError(t, err)

		assert.Equal(t, pkt, pkt2)
		assert.Equal(t, n1, n2)
	})
}

func TestUnsubscribeDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			32, // remaining length
			0,  // packet id
			7,
			0, // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // topic
			8,
			'/', 'a', '/', 'b', '/', '#', '/', 'c',
			0, // topic
			10,
			'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
		}

		pkt := NewUnsubscribe()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, 3, len(pkt.Topics))
		assert.Equal(t, "gomqtt", pkt.Topics[0])
		assert.Equal(t, "/a/b/#/c", pkt.Topics[1])
		assert.Equal(t, "/a/b/#/cdd", pkt.Topics[2])

		assertDecodeError(t, m, UNSUBSCRIBE, 4, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			2, // remaining length
			0, // packet id
			7,
			// < empty topic list
		})

		assertDecodeError(t, m, UNSUBSCRIBE, 2, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			6, // < wrong remaining length
			0, // packet id
			7,
		})

		assertDecodeError(t, m, UNSUBSCRIBE, 2, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			0, // remaining length
			// missing packet id
		})

		assertDecodeError(t, m, UNSUBSCRIBE, 6, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			10, // remaining length
			0,  // packet id
			7,
			0, // topic
			9, // < wrong size
			'g', 'o', 'm', 'q', 't', 't',
		})

		assertDecodeError(t, m, UNSUBSCRIBE, 4, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			10, // remaining length
			0,  // packet ID
			0,  // < zero packet id
			0,  // topic name
			6,
			'g', 'o', 'm', 'q', 't', 't',
		})
	})
}

func TestUnsubscribeEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			32, // remaining length
			0,  // packet id
			7,
			0, // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // topic
			8,
			'/', 'a', '/', 'b', '/', '#', '/', 'c',
			0, // topic
			10,
			'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
		}

		pkt := NewUnsubscribe()
		pkt.ID = 7
		pkt.Topics = []string{
			"gomqtt",
			"/a/b/#/c",
			"/a/b/#/cdd",
		}

		dst := make([]byte, 100)
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		// small buffer
		assertEncodeError(t, m, 1, 1, &Unsubscribe{
			ID:     7,
			Topics: []string{"gomqtt"},
		})

		assertEncodeError(t, m, 0, 6, &Unsubscribe{
			ID:     7,
			Topics: []string{string(make([]byte, 65536))}, // too big
		})

		assertEncodeError(t, m, 0, 2, &Unsubscribe{
			ID: 0, // < missing
		})
	})
}

func BenchmarkUnsubscribeEncode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		pkt := NewUnsubscribe()
		pkt.ID = 1
		pkt.Topics = []string{"t"}

		buf := make([]byte, pkt.Len(m))

		for i := 0; i < b.N; i++ {
			_, err := pkt.Encode(m, buf)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkUnsubscribeDecode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		packet := []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			5, // remaining length
			0, // packet id
			1,
			0, // topic
			1,
			't',
		}

		pkt := NewUnsubscribe()

		for i := 0; i < b.N; i++ {
			_, err := pkt.Decode(m, packet)
			if err != nil {
				panic(err)
			}
		}
	})
}
