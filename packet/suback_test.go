package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubackInterface(t *testing.T) {
	pkt := NewSuback()
	pkt.ReturnCodes = []QOS{0, 1}

	assert.Equal(t, pkt.Type(), SUBACK)
	assert.Equal(t, "<Suback ID=0 ReturnCodes=[0, 1]>", pkt.String())
}

func TestSubackDecode(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	pkt := NewSuback()
	n, err := pkt.Decode(packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, 4, len(pkt.ReturnCodes))
}

func TestSubackDecodeError1(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		1, // < wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
		0, // return code 1
	}

	pkt := NewSuback()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubackDecodeError2(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x81, // < wrong return code
	}

	pkt := NewSuback()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubackDecodeError3(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		1, // < wrong remaining length
		0, // packet ID MSB
	}

	pkt := NewSuback()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubackDecodeError4(t *testing.T) {
	packet := []byte{
		byte(PUBCOMP << 4), // < wrong packet type
		3,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // return code 1
	}

	pkt := NewSuback()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubackDecodeError5(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		3,
		0, // packet ID MSB
		0, // packet ID LSB < zero packet id
		0,
	}

	pkt := NewSuback()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubackEncode(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	pkt := NewSuback()
	pkt.ID = 7
	pkt.ReturnCodes = []QOS{0, 1, 2, 0x80}

	dst := make([]byte, 10)
	n, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, dst[:n])
}

func TestSubackEncodeError1(t *testing.T) {
	pkt := NewSuback()
	pkt.ID = 7
	pkt.ReturnCodes = []QOS{0x81}

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSubackEncodeError2(t *testing.T) {
	pkt := NewSuback()
	pkt.ID = 7
	pkt.ReturnCodes = []QOS{0x80}

	dst := make([]byte, pkt.Len()-1)
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSubackEncodeError3(t *testing.T) {
	pkt := NewSuback()
	pkt.ID = 0 // < zero packet id
	pkt.ReturnCodes = []QOS{0x80}

	dst := make([]byte, pkt.Len()-1)
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSubackEqualDecodeEncode(t *testing.T) {
	packet := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	pkt := NewSuback()
	n, err := pkt.Decode(packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)

	dst := make([]byte, 100)
	n2, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n2)
	assert.Equal(t, packet, dst[:n2])

	n3, err := pkt.Decode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n3)
}

func BenchmarkSubackEncode(b *testing.B) {
	pkt := NewSuback()
	pkt.ID = 1
	pkt.ReturnCodes = []QOS{0}

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSubackDecode(b *testing.B) {
	packet := []byte{
		byte(SUBACK << 4),
		3,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // return code 1
	}

	pkt := NewSuback()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(packet)
		if err != nil {
			panic(err)
		}
	}
}
