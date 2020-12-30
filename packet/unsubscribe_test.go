package packet

import "testing"

func TestUnsubscribe(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			28, // remaining length
			0,  // packet id
			7,
			0, // topic
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // topic
			8,
			'/', 'a', '/', 'b', '/', '*', '/', 'c',
			0, // topic
			6,
			'/', 'a', '/', 'b', '/', '#',
		}, &Unsubscribe{
			Topics: []string{
				"gomqtt",
				"/a/b/*/c",
				"/a/b/#",
			},
			ID: 7,
		}, `<Unsubscribe ID=7 Topics=["gomqtt", "/a/b/*/c", "/a/b/#"]>`)

		assertDecodeError(t, m, UNSUBSCRIBE, 4, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			2, // remaining length
			0, // packet id
			7,
			// < missing topic
		}, "missing topics")

		assertDecodeError(t, m, UNSUBSCRIBE, 2, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			6, // < wrong remaining length
			0, // packet id
			7,
		}, ErrRemainingLengthMismatch)

		assertDecodeError(t, m, UNSUBSCRIBE, 2, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			0, // remaining length
			// missing packet id
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, UNSUBSCRIBE, 4, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			2, // remaining length
			0, // packet ID
			0, // < zero packet id
		}, ErrInvalidPacketID)

		assertDecodeError(t, m, UNSUBSCRIBE, 4, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			2, // remaining length
			0, // packet id
			7,
		}, "missing topics")

		assertDecodeError(t, m, UNSUBSCRIBE, 6, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			4, // remaining length
			0, // packet id
			7,
			0, // topic
			0, // < zero topic
		}, "invalid topic")

		assertDecodeError(t, m, UNSUBSCRIBE, 6, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			10, // remaining length
			0,  // packet id
			7,
			0, // topic
			9, // < wrong size
			'g', 'o', 'm', 'q', 't', 't',
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, UNSUBSCRIBE, 12, []byte{
			byte(UNSUBSCRIBE<<4) | 2,
			11, // remaining length
			0,  // packet ID
			7,
			0, // topic name
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // < superfluous byte
		}, ErrInsufficientBufferSize)

		// small buffer
		assertEncodeError(t, m, 1, 1, &Unsubscribe{}, ErrInsufficientBufferSize)

		assertEncodeError(t, m, 0, 4, &Unsubscribe{
			ID:     7,
			Topics: []string{""}, // < zero empty topic
		}, "invalid topic")

		assertEncodeError(t, m, 0, 6, &Unsubscribe{
			ID:     7,
			Topics: []string{longString}, // too big
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 2, &Unsubscribe{
			ID: 0, // < missing
		}, ErrInvalidPacketID)
	})
}

func BenchmarkUnsubscribe(b *testing.B) {
	benchPacket(b, &Unsubscribe{
		Topics: []string{"t"},
		ID:     7,
	})
}
