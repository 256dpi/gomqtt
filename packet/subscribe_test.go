package packet

import "testing"

func TestSubscribe(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(SUBSCRIBE<<4) | 2,
			31, // remaining length
			0,  // packet id
			7,
			0, // topic name
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // qos
			0, // topic
			8,
			'/', 'a', '/', 'b', '/', '*', '/', 'c',
			1, // qos
			0, // topic
			6,
			'/', 'a', '/', 'b', '/', '#',
			2, // qos
		}, &Subscribe{
			Subscriptions: []Subscription{
				{Topic: "gomqtt", QOS: 0},
				{Topic: "/a/b/*/c", QOS: 1},
				{Topic: "/a/b/#", QOS: 2},
			},
			ID: 7,
		}, `<Subscribe ID=7 Subscriptions=["gomqtt"=>0, "/a/b/*/c"=>1, "/a/b/#"=>2]>`)

		assertDecodeError(t, m, SUBSCRIBE, 2, []byte{
			byte(SUBSCRIBE<<4) | 2,
			9, // < remaining length: too much
		}, ErrRemainingLengthMismatch)

		assertDecodeError(t, m, SUBSCRIBE, 2, []byte{
			byte(SUBSCRIBE<<4) | 2,
			0, // remaining length
			// < missing packet id
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, SUBSCRIBE, 4, []byte{
			byte(SUBSCRIBE<<4) | 2,
			2, // remaining length
			0, // packet id
			7,
			// < missing subscription
		}, "missing subscriptions")

		assertDecodeError(t, m, SUBSCRIBE, 6, []byte{
			byte(SUBSCRIBE<<4) | 2,
			4, // remaining length
			0, // packet id
			7,
			0, // topic
			0, // < zero topic
		}, ErrInvalidTopic)

		assertDecodeError(t, m, SUBSCRIBE, 6, []byte{
			byte(SUBSCRIBE<<4) | 2,
			5, // remaining length
			0, // packet id
			7,
			0, // topic
			2, // < wrong size
			's',
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, SUBSCRIBE, 7, []byte{
			byte(SUBSCRIBE<<4) | 2,
			5, // remaining length
			0, // packet id
			7,
			0, // topic
			1,
			's',
			// < missing qos
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, SUBSCRIBE, 4, []byte{
			byte(SUBSCRIBE<<4) | 2,
			6, // remaining length
			0, // packet id
			0, // < zero packet id
			0, // topic
			1,
			's',
			0,
		}, ErrInvalidPacketID)

		assertDecodeError(t, m, SUBSCRIBE, 8, []byte{
			byte(SUBSCRIBE<<4) | 2,
			6, // remaining length
			0, // packet id
			7,
			0, // topic
			1,
			's',
			0x81, // < invalid qos
		}, ErrInvalidQOSLevel)

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
		}, ErrInsufficientBufferSize)

		// small buffer
		assertEncodeError(t, m, 1, 1, &Subscribe{}, ErrInsufficientBufferSize)

		assertEncodeError(t, m, 0, 2, &Subscribe{
			ID: 0, // < missing
		}, ErrInvalidPacketID)

		assertEncodeError(t, m, 0, 4, &Subscribe{
			ID:            7,
			Subscriptions: []Subscription{},
		}, "missing subscriptions")

		assertEncodeError(t, m, 0, 4, &Subscribe{
			ID: 7,
			Subscriptions: []Subscription{
				{
					Topic: "", // < zero topic
				},
			},
		}, ErrInvalidTopic)

		assertEncodeError(t, m, 0, 6, &Subscribe{
			ID: 7,
			Subscriptions: []Subscription{
				{
					Topic: longString, // < too big
				},
			},
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 10, &Subscribe{
			ID: 7,
			Subscriptions: []Subscription{
				{
					Topic: "test",
					QOS:   0x81, // < invalid qos
				},
			},
		}, ErrInvalidQOSLevel)
	})
}

func BenchmarkSubscribe(b *testing.B) {
	benchPacket(b, &Subscribe{
		Subscriptions: []Subscription{
			{Topic: "topic", QOS: 0},
		},
		ID: 7,
	})
}
