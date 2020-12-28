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
			0, // session not present
			0, // connection accepted
		}

		pkt := NewConnack()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.False(t, pkt.SessionPresent)
		assert.Equal(t, ConnectionAccepted, pkt.ReturnCode)

		packet = []byte{
			byte(CONNACK << 4),
			3, // < remaining length: wrong size
			0, // session not present
			0, // connection accepted
		}

		pkt = NewConnack()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // session not present
			// < wrong packet size
		}

		pkt = NewConnack()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(CONNACK << 4),
			2,  // remaining length
			64, // < wrong value
			0,  // connection accepted
		}

		pkt = NewConnack()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0,
			6, // < wrong code
		}

		pkt = NewConnack()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)

		packet = []byte{
			byte(CONNACK << 4),
			1, // < wrong remaining length
			0,
			6,
		}

		pkt = NewConnack()
		_, err = pkt.Decode(m, packet)
		assert.Error(t, err)
	})
}

func TestConnackEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(CONNACK << 4),
			2, // remaining length
			1, // session present
			0, // connection accepted
		}

		pkt := NewConnack()
		pkt.ReturnCode = ConnectionAccepted
		pkt.SessionPresent = true

		dst := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, packet, dst[:n])

		pkt = NewConnack()
		dst = make([]byte, 3) // < wrong buffer size

		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 0, n)

		pkt = NewConnack()
		pkt.ReturnCode = 11 // < wrong return code

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.Error(t, err)
		assert.Equal(t, 3, n)
	})
}

func BenchmarkConnackEncode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		pkt := NewConnack()
		pkt.ReturnCode = ConnectionAccepted
		pkt.SessionPresent = true

		buf := make([]byte, pkt.Len(m))

		for i := 0; i < b.N; i++ {
			_, err := pkt.Encode(m, buf)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkConnackDecode(b *testing.B) {
	benchTest(b, func(b *testing.B, m Mode) {
		packet := []byte{
			byte(CONNACK << 4),
			2, // remaining length
			0, // session not present
			0, // connection accepted
		}

		pkt := NewConnack()

		for i := 0; i < b.N; i++ {
			_, err := pkt.Decode(m, packet)
			if err != nil {
				panic(err)
			}
		}
	})
}
