package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		pkt := NewSubscribe()
		pkt.ID = 1
		pkt.Subscriptions = []Subscription{
			{Topic: "foo", QOS: QOSAtMostOnce},
			{Topic: "bar", QOS: QOSAtLeastOnce},
		}

		assert.Equal(t, pkt.Type(), SUBSCRIBE)
		assert.Equal(t, `<Subscribe ID=1 Subscriptions=["foo"=>0, "bar"=>1]>`, pkt.String())

		buf := make([]byte, pkt.Len(m))
		n1, err := pkt.Encode(m, buf)
		assert.NoError(t, err)

		pkt2 := NewSubscribe()
		n2, err := pkt2.Decode(m, buf)
		assert.NoError(t, err)

		assert.Equal(t, pkt, pkt2)
		assert.Equal(t, n1, n2)
	})
}

func TestSubscribeDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(SUBSCRIBE<<4) | 2,
			35, // remaining length
			0,  // packet id
			7,
			0, // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // qos
			0, // topic
			8,
			'/', 'a', '/', 'b', '/', '#', '/', 'c',
			1, // qos
			0, // topic
			10,
			'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
			2, // qos
		}

		pkt := NewSubscribe()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, 3, len(pkt.Subscriptions))
		assert.Equal(t, "gomqtt", pkt.Subscriptions[0].Topic)
		assert.Equal(t, QOS(0), pkt.Subscriptions[0].QOS)
		assert.Equal(t, "/a/b/#/c", pkt.Subscriptions[1].Topic)
		assert.Equal(t, QOS(1), pkt.Subscriptions[1].QOS)
		assert.Equal(t, "/a/b/#/cdd", pkt.Subscriptions[2].Topic)
		assert.Equal(t, QOS(2), pkt.Subscriptions[2].QOS)

		assertDecodeError(t, m, SUBSCRIBE, 2, []byte{
			byte(SUBSCRIBE<<4) | 2,
			9, // < remaining length: too much
		})

		assertDecodeError(t, m, SUBSCRIBE, 2, []byte{
			byte(SUBSCRIBE<<4) | 2,
			0, // remaining length
			// < missing packet id
		})

		assertDecodeError(t, m, SUBSCRIBE, 4, []byte{
			byte(SUBSCRIBE<<4) | 2,
			2, // remaining length
			0, // packet id
			7,
			// < missing subscription
		})

		assertDecodeError(t, m, SUBSCRIBE, 6, []byte{
			byte(SUBSCRIBE<<4) | 2,
			5, // remaining length
			0, // packet id
			7,
			0, // topic
			2, // < wrong size
			's',
		})

		assertDecodeError(t, m, SUBSCRIBE, 7, []byte{
			byte(SUBSCRIBE<<4) | 2,
			5, // remaining length
			0, // packet id
			7,
			0, // topic
			1,
			's',
			// < missing qos
		})

		assertDecodeError(t, m, SUBSCRIBE, 4, []byte{
			byte(SUBSCRIBE<<4) | 2,
			6, // remaining length
			0, // packet id
			0, // < zero packet id
			0, // topic
			1,
			's',
			0,
		})

		assertDecodeError(t, m, SUBSCRIBE, 8, []byte{
			byte(SUBSCRIBE<<4) | 2,
			6, // remaining length
			0, // packet id
			7,
			0, // topic
			1,
			's',
			0x81, // < invalid qos
		})

		assertDecodeError(t, m, SUBSCRIBE, 8, []byte{
			byte(SUBSCRIBE<<4) | 2,
			7, // remaining length
			0, // packet id
			7,
			0, // topic
			1,
			's',
			0, // qos
			0, // < superfluous byte
		})
	})
}

func TestSubscribeEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(SUBSCRIBE<<4) | 2,
			35, // remaining length
			0,  // packet id
			7,
			0, // topic name
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // qos
			0, // topic
			8,
			'/', 'a', '/', 'b', '/', '#', '/', 'c',
			1, // qos
			0, // topic
			10,
			'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
			2, // qos
		}

		pkt := NewSubscribe()
		pkt.ID = 7
		pkt.Subscriptions = []Subscription{
			{Topic: "gomqtt", QOS: 0},
			{Topic: "/a/b/#/c", QOS: 1},
			{Topic: "/a/b/#/cdd", QOS: 2},
		}

		dst := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst)

		// small buffer
		assertEncodeError(t, m, 1, 1, &Subscribe{})

		assertEncodeError(t, m, 0, 2, &Subscribe{
			ID: 0, // < missing
		})

		assertEncodeError(t, m, 0, 6, &Subscribe{
			ID: 7,
			Subscriptions: []Subscription{
				{
					Topic: longString, // < too big
					QOS:   0,
				},
			},
		})

		assertEncodeError(t, m, 0, 16, &Subscribe{
			ID: 7,
			Subscriptions: []Subscription{
				{
					Topic: string(make([]byte, 10)),
					QOS:   0x81, // < invalid qos
				},
			},
		})
	})
}

func BenchmarkSubscribe(b *testing.B) {
	benchPacket(b, &Subscribe{
		Subscriptions: []Subscription{
			{Topic: "t", QOS: 0},
		},
		ID: 7,
	})
}
