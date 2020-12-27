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
			32,
			0, // packet ID MSB
			7, // packet ID LSB
			0, // topic name MSB
			6, // topic name LSB
			'g', 'o', 'm', 'q', 't', 't',
			0, // topic name MSB
			8, // topic name LSB
			'/', 'a', '/', 'b', '/', '#', '/', 'c',
			0,  // topic name MSB
			10, // topic name LSB
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

		packet = []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			2,
			0, // packet ID MSB
			7, // packet ID LSB
			// empty topic list
		}

		pkt = NewUnsubscribe()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			6, // < wrong remaining length
			0, // packet ID MSB
			7, // packet ID LSB
		}

		pkt = NewUnsubscribe()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			0,
			// missing packet id
		}

		pkt = NewUnsubscribe()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			10,
			0, // packet ID MSB
			7, // packet ID LSB
			0, // topic name MSB
			9, // topic name LSB < wrong size
			'g', 'o', 'm', 'q', 't', 't',
		}

		pkt = NewUnsubscribe()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			10,
			0, // packet ID MSB
			0, // packet ID LSB < zero packet id
			0, // topic name MSB
			6, // topic name LSB
			'g', 'o', 'm', 'q', 't', 't',
		}

		pkt = NewUnsubscribe()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)
	})
}

func TestUnsubscribeEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			32,
			0, // packet ID MSB
			7, // packet ID LSB
			0, // topic name MSB
			6, // topic name LSB
			'g', 'o', 'm', 'q', 't', 't',
			0, // topic name MSB
			8, // topic name LSB
			'/', 'a', '/', 'b', '/', '#', '/', 'c',
			0,  // topic name MSB
			10, // topic name LSB
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

		pkt = NewUnsubscribe()
		pkt.ID = 7
		pkt.Topics = []string{"gomqtt"}

		dst = make([]byte, 1) // < too small
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 0, n)

		pkt = NewUnsubscribe()
		pkt.ID = 7
		pkt.Topics = []string{string(make([]byte, 65536))}

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 6, n)

		pkt = NewUnsubscribe()
		pkt.ID = 0 // < zero packet id

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 0, n)
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
			5,
			0, // packet ID MSB
			1, // packet ID LSB
			0, // topic name MSB
			1, // topic name LSB
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
