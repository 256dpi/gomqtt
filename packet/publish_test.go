package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		pkt := NewPublish()
		pkt.ID = 1
		pkt.Message.Topic = "foo"
		pkt.Message.QOS = 2
		pkt.Dup = true

		assert.Equal(t, pkt.Type(), PUBLISH)
		assert.Equal(t, `<Publish ID=1 Message=<Message Topic="foo" QOS=2 Retain=false Payload=> Dup=true>`, pkt.String())

		buf := make([]byte, pkt.Len(m))
		n1, err := pkt.Encode(m, buf)
		assert.NoError(t, err)

		pkt2 := NewPublish()
		n2, err := pkt2.Decode(m, buf)
		assert.NoError(t, err)

		assert.Equal(t, pkt, pkt2)
		assert.Equal(t, n1, n2)
	})
}

func TestPublishDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(PUBLISH<<4) | 11,
			22, // remaining length
			0,  // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // packet id
			7,
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		}

		pkt := NewPublish()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, ID(7), pkt.ID)
		assert.Equal(t, "gomqtt", pkt.Message.Topic)
		assert.Equal(t, []byte("send me home"), pkt.Message.Payload)
		assert.Equal(t, QOS(1), pkt.Message.QOS)
		assert.Equal(t, true, pkt.Message.Retain)
		assert.Equal(t, true, pkt.Dup)

		packet = []byte{
			byte(PUBLISH << 4),
			20, // remaining length
			0,  // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		}

		pkt = NewPublish()
		n, err = pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, ID(0), pkt.ID)
		assert.Equal(t, "gomqtt", pkt.Message.Topic)
		assert.Equal(t, []byte("send me home"), pkt.Message.Payload)
		assert.Equal(t, QOS(0), pkt.Message.QOS)
		assert.Equal(t, false, pkt.Message.Retain)
		assert.Equal(t, false, pkt.Dup)

		assertDecodeError(t, m, PUBLISH, []byte{
			byte(PUBLISH << 4),
			2, // < remaining length: too much
		})

		assertDecodeError(t, m, PUBLISH, []byte{
			byte(PUBLISH<<4) | 6, // < wrong qos
			0,                    // remaining length
		})

		assertDecodeError(t, m, PUBLISH, []byte{
			byte(PUBLISH << 4),
			0, // remaining length
			// < missing topic stuff
		})

		assertDecodeError(t, m, PUBLISH, []byte{
			byte(PUBLISH << 4),
			2, // remaining length
			0, // topic
			1,
			// < missing topic string
		})

		assertDecodeError(t, m, PUBLISH, []byte{
			byte(PUBLISH<<4) | 2,
			2, // remaining length
			0, // topic
			1,
			't',
			// < missing packet id
		})

		assertDecodeError(t, m, PUBLISH, []byte{
			byte(PUBLISH<<4) | 2,
			2, // remaining length
			0, // topic
			1,
			't',
			0,
			0, // < zero packet id
		})
	})
}

func TestPublishEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(PUBLISH<<4) | 11,
			22, // remaining length
			0,  // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // packet id
			7,
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		}

		pkt := NewPublish()
		pkt.Message.Topic = "gomqtt"
		pkt.Message.QOS = QOSAtLeastOnce
		pkt.Message.Retain = true
		pkt.Dup = true
		pkt.ID = 7
		pkt.Message.Payload = []byte("send me home")

		dst := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		packet = []byte{
			byte(PUBLISH << 4),
			20, // remaining length
			0,  // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		}

		pkt = NewPublish()
		pkt.Message.Topic = "gomqtt"
		pkt.Message.Payload = []byte("send me home")

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		pkt = NewPublish()
		pkt.Message.Topic = "" // < empty topic

		dst = make([]byte, pkt.Len(m))
		_, err = pkt.Encode(m, dst)
		assert.Error(t, err)

		pkt = NewPublish()
		pkt.Message.Topic = "t"
		pkt.Message.QOS = 3 // < wrong qos

		dst = make([]byte, pkt.Len(m))
		_, err = pkt.Encode(m, dst)
		assert.Error(t, err)

		pkt = NewPublish()
		pkt.Message.Topic = "t"

		dst = make([]byte, 1) // < too small
		_, err = pkt.Encode(m, dst)
		assert.Error(t, err)

		pkt = NewPublish()
		pkt.Message.Topic = string(make([]byte, 65536)) // < too big

		dst = make([]byte, pkt.Len(m))
		_, err = pkt.Encode(m, dst)
		assert.Error(t, err)

		pkt = NewPublish()
		pkt.Message.Topic = "test"
		pkt.Message.QOS = 1
		pkt.ID = 0 // < zero packet id

		dst = make([]byte, pkt.Len(m))
		_, err = pkt.Encode(m, dst)
		assert.Error(t, err)
	})
}

func BenchmarkPublishEncode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		pkt := NewPublish()
		pkt.Message.Topic = "t"
		pkt.Message.QOS = QOSAtLeastOnce
		pkt.ID = 1
		pkt.Message.Payload = []byte("p")

		buf := make([]byte, pkt.Len(m))

		for i := 0; i < b.N; i++ {
			_, err := pkt.Encode(m, buf)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkPublishDecode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		packet := []byte{
			byte(PUBLISH<<4) | 2,
			6, // remaining length
			0, // topic
			1,
			't',
			0, // packet id
			1,
			'p',
		}

		pkt := NewPublish()

		for i := 0; i < b.N; i++ {
			_, err := pkt.Decode(m, packet)
			if err != nil {
				panic(err)
			}
		}
	})
}
