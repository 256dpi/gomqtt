package packet

import "testing"

func TestSuback(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(SUBACK << 4),
			6, // remaining length
			0, // packet id
			7,
			0,    // return code 1
			1,    // return code 2
			2,    // return code 3
			0x80, // return code 4
		}, &Suback{
			ReturnCodes: []QOS{0, 1, 2, 0x80},
			ID:          7,
		}, `<Suback ID=7 ReturnCodes=[0, 1, 2, 128]>`)

		assertDecodeError(t, m, SUBACK, 2, []byte{
			byte(SUBACK << 4),
			1, // < wrong remaining length
			0, // packet id
			7,
			0, // return code 1
		}, ErrRemainingLengthMismatch)

		assertDecodeError(t, m, SUBACK, 8, []byte{
			byte(SUBACK << 4),
			6, // remaining length
			0, // packet id
			7,
			0,    // return code 1
			1,    // return code 2
			2,    // return code 3
			0x81, // < wrong return code
		}, "invalid return code")

		assertDecodeError(t, m, SUBACK, 2, []byte{
			byte(SUBACK << 4),
			1, // remaining length
			0, // packet id
			// < missing packet id
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, SUBACK, 1, []byte{
			byte(PUBCOMP << 4), // < wrong packet type
			3,                  // remaining length
			0,                  // packet id
			7,
			0, // return code 1
		}, ErrInvalidPacketType)

		assertDecodeError(t, m, SUBACK, 4, []byte{
			byte(SUBACK << 4),
			3, // remaining length
			0, // packet id
			0, // < zero packet id
			0,
		}, ErrInvalidPacketID)

		// small buffer
		assertEncodeError(t, m, 1, 1, &Suback{}, ErrInsufficientBufferSize)

		assertEncodeError(t, m, 0, 4, &Suback{
			ID:          7,
			ReturnCodes: []QOS{0x81}, // < invalid
		}, "invalid return code")

		assertEncodeError(t, m, 0, 2, &Suback{
			ID:          0, // < zero packet id
			ReturnCodes: []QOS{0},
		}, ErrInvalidPacketID)
	})
}

func BenchmarkSuback(b *testing.B) {
	benchPacket(b, &Suback{
		ReturnCodes: []QOS{0},
		ID:          7,
	})
}
