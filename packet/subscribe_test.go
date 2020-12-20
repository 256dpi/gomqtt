package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeInterface(t *testing.T) {
	pkt := NewSubscribe()
	pkt.Subscriptions = []Subscription{
		{Topic: "foo", QOS: QOSAtMostOnce},
		{Topic: "bar", QOS: QOSAtLeastOnce},
	}

	assert.Equal(t, pkt.Type(), SUBSCRIBE)
	assert.Equal(t, "<Subscribe ID=0 Subscriptions=[\"foo\"=>0, \"bar\"=>1]>", pkt.String())
}

func TestSubscribeDecode(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		35,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // QOS
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		1,  // QOS
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
		2, // QOS
	}

	pkt := NewSubscribe()
	n, err := pkt.Decode(packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, 3, len(pkt.Subscriptions))
	assert.Equal(t, "gomqtt", pkt.Subscriptions[0].Topic)
	assert.Equal(t, QOS(0), pkt.Subscriptions[0].QOS)
	assert.Equal(t, "/a/b/#/c", pkt.Subscriptions[1].Topic)
	assert.Equal(t, QOS(1), pkt.Subscriptions[1].QOS)
	assert.Equal(t, "/a/b/#/cdd", pkt.Subscriptions[2].Topic)
	assert.Equal(t, QOS(2), pkt.Subscriptions[2].QOS)
}

func TestSubscribeDecodeError1(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		9, // < too much
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeDecodeError2(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		0,
		// < missing packet id
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeDecodeError3(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		2,
		0, // packet ID MSB
		7, // packet ID LSB
		// < missing subscription
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeDecodeError4(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		2, // topic name LSB < wrong size
		's',
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeDecodeError5(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		's',
		// < missing qos
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeDecodeError6(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		0, // packet ID LSB < zero packet id
		0, // topic name MSB
		1, // topic name LSB
		's',
		0,
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeDecodeError7(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		's',
		0x81, // < invalid qos
	}

	pkt := NewSubscribe()
	_, err := pkt.Decode(packet)

	assert.Error(t, err)
}

func TestSubscribeEncode(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		35,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // QOS
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		1,  // QOS
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
		2, // QOS
	}

	pkt := NewSubscribe()
	pkt.ID = 7
	pkt.Subscriptions = []Subscription{
		{Topic: "gomqtt", QOS: 0},
		{Topic: "/a/b/#/c", QOS: 1},
		{Topic: "/a/b/#/cdd", QOS: 2},
	}

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, dst)
}

func TestSubscribeEncodeError1(t *testing.T) {
	pkt := NewSubscribe()
	pkt.ID = 7

	dst := make([]byte, 1) // < too small
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestSubscribeEncodeError2(t *testing.T) {
	pkt := NewSubscribe()
	pkt.ID = 7
	pkt.Subscriptions = []Subscription{
		{Topic: string(make([]byte, 65536)), QOS: 0}, // too big
	}

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestSubscribeEncodeError3(t *testing.T) {
	pkt := NewSubscribe()
	pkt.ID = 0 // < zero packet id

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestSubscribeEncodeError4(t *testing.T) {
	pkt := NewSubscribe()
	pkt.ID = 7
	pkt.Subscriptions = []Subscription{
		{Topic: string(make([]byte, 10)), QOS: 0x81}, // invalid qos
	}

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestSubscribeEqualDecodeEncode(t *testing.T) {
	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		35,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // QOS
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		1,  // QOS
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
		2, // QOS
	}

	pkt := NewSubscribe()
	n, err := pkt.Decode(packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)

	dst := make([]byte, pkt.Len())
	n2, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n2)
	assert.Equal(t, packet, dst[:n2])

	n3, err := pkt.Decode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n3)
}

func BenchmarkSubscribeEncode(b *testing.B) {
	b.ReportAllocs()

	pkt := NewSubscribe()
	pkt.ID = 7
	pkt.Subscriptions = []Subscription{
		{Topic: "t", QOS: 0},
	}

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSubscribeDecode(b *testing.B) {
	b.ReportAllocs()

	packet := []byte{
		byte(SUBSCRIBE<<4) | 2,
		6,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		't',
		0, // QOS
	}

	pkt := NewSubscribe()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(packet)
		if err != nil {
			panic(err)
		}
	}
}
