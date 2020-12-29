package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnackReturnCodes(t *testing.T) {
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
		pkt := NewConnack()
		pkt.SessionPresent = true
		pkt.ReturnCode = BadUsernameOrPassword

		assert.Equal(t, pkt.Type(), CONNACK)
		assert.Equal(t, "<Connack SessionPresent=true ReturnCode=4>", pkt.String())

		buf := make([]byte, pkt.Len(m))
		n1, err := pkt.Encode(m, buf)
		assert.NoError(t, err)

		pkt2 := NewConnack()
		n2, err := pkt2.Decode(m, buf)
		assert.NoError(t, err)

		assert.Equal(t, pkt, pkt2)
		assert.Equal(t, n1, n2)
	})
}

func TestConnackDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // session present
			0, // return code
		}

		pkt := NewConnack()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.False(t, pkt.SessionPresent)
		assert.Equal(t, ConnectionAccepted, pkt.ReturnCode)

		assertDecodeError(t, m, CONNACK, 2, []byte{
			byte(CONNACK << 4),
			3, // < remaining length: wrong size
			0, // session present
			0, // return code
		})

		assertDecodeError(t, m, CONNACK, 2, []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // session present
			// < wrong packet size
		})

		assertDecodeError(t, m, CONNACK, 3, []byte{
			byte(CONNACK << 4),
			2,  // remaining length
			64, // < wrong value
			0,  // return code
		})

		assertDecodeError(t, m, CONNACK, 4, []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // session present
			6, // < wrong return code
		})

		assertDecodeError(t, m, CONNACK, 2, []byte{
			byte(CONNACK << 4),
			1, // < wrong remaining length
			0, // session present
			6, // return code
		})

		assertDecodeError(t, m, CONNACK, 4, []byte{
			byte(CONNACK << 4),
			3, // remaining length
			0, // session present
			6, // return code
			0, // < superfluous byte
		})
	})
}

func TestConnackEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(CONNACK << 4),
			2, // remaining length
			1, // session present
			0, // return code
		}

		pkt := NewConnack()
		pkt.ReturnCode = ConnectionAccepted
		pkt.SessionPresent = true

		dst := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, packet, dst[:n])

		// small buffer
		assertEncodeError(t, m, 1, 1, &Connack{})

		assertEncodeError(t, m, 0, 3, &Connack{
			ReturnCode: 11, // < wrong return code
		})
	})
}

func BenchmarkConnack(b *testing.B) {
	benchPacket(b, &Connack{
		ReturnCode:     ConnectionAccepted,
		SessionPresent: true,
	})
}
