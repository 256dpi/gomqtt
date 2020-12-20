package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsubscribeInterface(t *testing.T) {
	pkt := NewUnsubscribe()
	pkt.Topics = []string{"foo", "bar"}

	assert.Equal(t, pkt.Type(), UNSUBSCRIBE)
	assert.Equal(t, "<Unsubscribe Topics=[\"foo\", \"bar\"]>", pkt.String())
}

func TestUnsubscribeDecode(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		32,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	pkt := NewUnsubscribe()
	n, err := pkt.Decode(M4, packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, 3, len(pkt.Topics))
	assert.Equal(t, "gomqtt", pkt.Topics[0])
	assert.Equal(t, "/a/b/#/c", pkt.Topics[1])
	assert.Equal(t, "/a/b/#/cdd", pkt.Topics[2])
}

func TestUnsubscribeDecodeError1(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		2,
		0, // packet ID MSB
		7, // packet ID LSB
		// empty topic list
	}

	pkt := NewUnsubscribe()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestUnsubscribeDecodeError2(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		6, // < wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
	}

	pkt := NewUnsubscribe()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestUnsubscribeDecodeError3(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		0,
		// missing packet id
	}

	pkt := NewUnsubscribe()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestUnsubscribeDecodeError4(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		10,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		9, // topic name LSB < wrong size
		'g', 'o', 'm', 'q', 't', 't',
	}

	pkt := NewUnsubscribe()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestUnsubscribeDecodeError5(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		10,
		0, // packet ID MSB
		0, // packet ID LSB < zero packet id
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
	}

	pkt := NewUnsubscribe()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestUnsubscribeEncode(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		32,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	pkt := NewUnsubscribe()
	pkt.ID = 7
	pkt.Topics = []string{
		"gomqtt",
		"/a/b/#/c",
		"/a/b/#/cdd",
	}

	dst := make([]byte, 100)
	n, err := pkt.Encode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, dst[:n])
}

func TestUnsubscribeEncodeError1(t *testing.T) {
	pkt := NewUnsubscribe()
	pkt.ID = 7
	pkt.Topics = []string{"gomqtt"}

	dst := make([]byte, 1) // < too small
	n, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestUnsubscribeEncodeError2(t *testing.T) {
	pkt := NewUnsubscribe()
	pkt.ID = 7
	pkt.Topics = []string{string(make([]byte, 65536))}

	dst := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
	assert.Equal(t, 6, n)
}

func TestUnsubscribeEncodeError3(t *testing.T) {
	pkt := NewUnsubscribe()
	pkt.ID = 0 // < zero packet id

	dst := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestUnsubscribeEqualDecodeEncode(t *testing.T) {
	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		32,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	pkt := NewUnsubscribe()
	n, err := pkt.Decode(M4, packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)

	dst := make([]byte, 100)
	n2, err := pkt.Encode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n2)
	assert.Equal(t, packet, dst[:n2])

	n3, err := pkt.Decode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n3)
}

func BenchmarkUnsubscribeEncode(b *testing.B) {
	b.ReportAllocs()

	pkt := NewUnsubscribe()
	pkt.ID = 1
	pkt.Topics = []string{"t"}

	buf := make([]byte, pkt.Len(M4))

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(M4, buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkUnsubscribeDecode(b *testing.B) {
	b.ReportAllocs()

	packet := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		't',
	}

	pkt := NewUnsubscribe()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(M4, packet)
		if err != nil {
			panic(err)
		}
	}
}
