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

func TestConnackInterface(t *testing.T) {
	pkt := NewConnack()

	assert.Equal(t, pkt.Type(), CONNACK)
	assert.Equal(t, "<Connack SessionPresent=false ReturnCode=0>", pkt.String())
}

func TestConnackDecode(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnack()

	n, err := pkt.Decode(M4, packet)

	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.False(t, pkt.SessionPresent)
	assert.Equal(t, ConnectionAccepted, pkt.ReturnCode)
}

func TestConnackDecodeError1(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		3, // < wrong size
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnack()

	_, err := pkt.Decode(M4, packet)
	assert.Error(t, err)
}

func TestConnackDecodeError2(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		// < wrong packet size
	}

	pkt := NewConnack()

	_, err := pkt.Decode(M4, packet)
	assert.Error(t, err)
}

func TestConnackDecodeError3(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		2,
		64, // < wrong value
		0,  // connection accepted
	}

	pkt := NewConnack()

	_, err := pkt.Decode(M4, packet)
	assert.Error(t, err)
}

func TestConnackDecodeError4(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		2,
		0,
		6, // < wrong code
	}

	pkt := NewConnack()

	_, err := pkt.Decode(M4, packet)
	assert.Error(t, err)
}

func TestConnackDecodeError5(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		1, // < wrong remaining length
		0,
		6,
	}

	pkt := NewConnack()

	_, err := pkt.Decode(M4, packet)
	assert.Error(t, err)
}

func TestConnackEncode(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		2,
		1, // session present
		0, // connection accepted
	}

	pkt := NewConnack()
	pkt.ReturnCode = ConnectionAccepted
	pkt.SessionPresent = true

	dst := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, packet, dst[:n])
}

func TestConnackEncodeError1(t *testing.T) {
	pkt := NewConnack()

	dst := make([]byte, 3) // < wrong buffer size
	n, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestConnackEncodeError2(t *testing.T) {
	pkt := NewConnack()
	pkt.ReturnCode = 11 // < wrong return code

	dst := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
	assert.Equal(t, 3, n)
}

func TestConnackEqualDecodeEncode(t *testing.T) {
	packet := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnack()
	n, err := pkt.Decode(M4, packet)

	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	dst := make([]byte, pkt.Len(M4))
	n2, err := pkt.Encode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, 4, n2)
	assert.Equal(t, packet, dst[:n2])

	n3, err := pkt.Decode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, 4, n3)
}

func BenchmarkConnackEncode(b *testing.B) {
	b.ReportAllocs()

	pkt := NewConnack()
	pkt.ReturnCode = ConnectionAccepted
	pkt.SessionPresent = true

	buf := make([]byte, pkt.Len(M4))

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(M4, buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkConnackDecode(b *testing.B) {
	b.ReportAllocs()

	packet := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnack()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(M4, packet)
		if err != nil {
			panic(err)
		}
	}
}
