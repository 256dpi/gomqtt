package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnackCode(t *testing.T) {
	assert.Equal(t, ConnectionAccepted.String(), ConnackCode(0).String())
	assert.Equal(t, InvalidProtocolVersion.String(), ConnackCode(1).String())
	assert.Equal(t, IdentifierRejected.String(), ConnackCode(2).String())
	assert.Equal(t, ServerUnavailable.String(), ConnackCode(3).String())
	assert.Equal(t, BadUsernameOrPassword.String(), ConnackCode(4).String())
	assert.Equal(t, NotAuthorized.String(), ConnackCode(5).String())
	assert.Equal(t, "invalid connack code", ConnackCode(6).String())
}

func TestConnack(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(CONNACK << 4),
			2, // remaining length
			1, // connack flags
			0, // return code
		}, &Connack{
			SessionPresent: true,
			ReturnCode:     ConnectionAccepted,
		}, "<Connack SessionPresent=true ReturnCode=0>")

		assertEncodeDecode(t, m, []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // connack flags
			1, // return code
		}, &Connack{
			SessionPresent: false,
			ReturnCode:     InvalidProtocolVersion,
		}, "<Connack SessionPresent=false ReturnCode=1>")

		assertDecodeError(t, m, CONNACK, 2, []byte{
			byte(CONNACK << 4),
			3, // < remaining length: wrong size
			0, // connack flags
			0, // return code
		}, ErrRemainingLengthMismatch)

		assertDecodeError(t, m, CONNACK, 3, []byte{
			byte(CONNACK << 4),
			1, // remaining length
			0, // connack flags
			// < missing return code
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, CONNACK, 3, []byte{
			byte(CONNACK << 4),
			2,  // remaining length
			64, // < wrong value
			0,  // return code
		}, "invalid connack flags")

		assertDecodeError(t, m, CONNACK, 4, []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // connack flags
			6, // < wrong return code
		}, "invalid return code")

		assertDecodeError(t, m, CONNACK, 4, []byte{
			byte(CONNACK << 4),
			3, // remaining length
			0, // connack flags
			0, // return code
			0, // < superfluous byte
		}, "leftover buffer length")

		// small buffer
		assertEncodeError(t, m, 1, 1, &Connack{}, ErrInsufficientBufferSize)

		assertEncodeError(t, m, 0, 3, &Connack{
			ReturnCode: 11, // < wrong return code
		}, "invalid return code")
	})
}

func BenchmarkConnack(b *testing.B) {
	benchPacket(b, &Connack{
		ReturnCode:     ConnectionAccepted,
		SessionPresent: true,
	})
}
