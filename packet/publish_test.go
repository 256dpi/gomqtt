package packet

import "testing"

func TestPublish(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(PUBLISH << 4),
			13, // remaining length
			0,  // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			'h', 'e', 'l', 'l', 'o',
		}, &Publish{
			Message: Message{
				Topic:          "gomqtt",
				Payload:        []byte("hello"),
				ScratchPayload: [255]byte{'h', 'e', 'l', 'l', 'o'},
			},
		}, `<Publish ID=0 Message=<Message Topic="gomqtt" QOS=0 Retain=false Payload=68656c6c6f> Dup=false>`)

		assertEncodeDecode(t, m, []byte{
			byte(PUBLISH<<4) | 11,
			15, // remaining length
			0,  // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // packet id
			7,
			'h', 'e', 'l', 'l', 'o',
		}, &Publish{
			Message: Message{
				Topic:          "gomqtt",
				Payload:        []byte("hello"),
				QOS:            QOSAtLeastOnce,
				Retain:         true,
				ScratchPayload: [255]byte{'h', 'e', 'l', 'l', 'o'},
			},
			Dup: true,
			ID:  7,
		}, `<Publish ID=7 Message=<Message Topic="gomqtt" QOS=1 Retain=true Payload=68656c6c6f> Dup=true>`)

		assertDecodeError(t, m, PUBLISH, 2, []byte{
			byte(PUBLISH << 4),
			1, // < wrong remaining length
		}, ErrRemainingLengthMismatch)

		assertDecodeError(t, m, PUBLISH, 2, []byte{
			byte(PUBLISH<<4) | 6, // < wrong qos
			0,                    // remaining length
		}, ErrInvalidQOSLevel)

		assertDecodeError(t, m, PUBLISH, 2, []byte{
			byte(PUBLISH << 4),
			0, // remaining length
			// < missing topic length
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, PUBLISH, 4, []byte{
			byte(PUBLISH << 4),
			2, // remaining length
			0, // topic
			0, // < zero topic
		}, "invalid topic")

		assertDecodeError(t, m, PUBLISH, 4, []byte{
			byte(PUBLISH << 4),
			2, // remaining length
			0, // topic
			1,
			// < missing topic string
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, PUBLISH, 5, []byte{
			byte(PUBLISH<<4) | 2,
			3, // remaining length
			0, // topic
			1,
			't',
			// < missing packet id
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, PUBLISH, 7, []byte{
			byte(PUBLISH<<4) | 2,
			5, // remaining length
			0, // topic
			1,
			't',
			0,
			0, // < zero packet id
		}, ErrInvalidPacketID)

		// small buffer
		assertEncodeError(t, m, 1, 1, &Publish{}, ErrInsufficientBufferSize)

		assertEncodeError(t, m, 0, 2, &Publish{
			Message: Message{
				Topic: "", // < missing
			},
		}, "invalid topic")

		assertEncodeError(t, m, 0, 0, &Publish{
			Message: Message{
				Topic: "test",
				QOS:   3, // < invalid
			},
		}, ErrInvalidQOSLevel)

		assertEncodeError(t, m, 0, 4, &Publish{
			Message: Message{
				Topic: longString, // < too big
			},
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 8, &Publish{
			ID: 0, // < missing
			Message: Message{
				Topic: "test",
				QOS:   1,
			},
		}, ErrInvalidPacketID)
	})
}

func BenchmarkPublish(b *testing.B) {
	benchPacket(b, &Publish{
		Message: Message{
			Topic:   "topic",
			QOS:     QOSAtLeastOnce,
			Payload: []byte("payload"),
			Retain:  true,
		},
		Dup: false,
		ID:  7,
	})
}
