package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	pkt := NewPublish()
	pkt.ID = 1
	pkt.Message.Topic = "foo"
	pkt.Message.QOS = 2
	pkt.Dup = true

	assert.Equal(t, pkt.Type(), PUBLISH)
	assert.Equal(t, `<Publish ID=1 Message=<Message Topic="foo" QOS=2 Retain=false Payload=> Dup=true>`, pkt.String())

	buf := make([]byte, pkt.Len(M4))
	n1, err := pkt.Encode(M4, buf)
	assert.NoError(t, err)

	pkt2 := NewPublish()
	n2, err := pkt2.Decode(M4, buf)
	assert.NoError(t, err)

	assert.Equal(t, pkt, pkt2)
	assert.Equal(t, n1, n2)
}

func TestPublishDecode1(t *testing.T) {
	packet := []byte{
		byte(PUBLISH<<4) | 11,
		22,
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublish()
	n, err := pkt.Decode(M4, packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, ID(7), pkt.ID)
	assert.Equal(t, "gomqtt", pkt.Message.Topic)
	assert.Equal(t, []byte("send me home"), pkt.Message.Payload)
	assert.Equal(t, QOS(1), pkt.Message.QOS)
	assert.Equal(t, true, pkt.Message.Retain)
	assert.Equal(t, true, pkt.Dup)
}

func TestPublishDecode2(t *testing.T) {
	packet := []byte{
		byte(PUBLISH << 4),
		20,
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublish()
	n, err := pkt.Decode(M4, packet)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, ID(0), pkt.ID)
	assert.Equal(t, "gomqtt", pkt.Message.Topic)
	assert.Equal(t, []byte("send me home"), pkt.Message.Payload)
	assert.Equal(t, QOS(0), pkt.Message.QOS)
	assert.Equal(t, false, pkt.Message.Retain)
	assert.Equal(t, false, pkt.Dup)
}

func TestPublishDecodeError1(t *testing.T) {
	packet := []byte{
		byte(PUBLISH << 4),
		2, // < too much
	}

	pkt := NewPublish()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestPublishDecodeError2(t *testing.T) {
	packet := []byte{
		byte(PUBLISH<<4) | 6, // < wrong qos
		0,
	}

	pkt := NewPublish()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestPublishDecodeError3(t *testing.T) {
	packet := []byte{
		byte(PUBLISH << 4),
		0,
		// < missing topic stuff
	}

	pkt := NewPublish()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestPublishDecodeError4(t *testing.T) {
	packet := []byte{
		byte(PUBLISH << 4),
		2,
		0, // topic name MSB
		1, // topic name LSB
		// < missing topic string
	}

	pkt := NewPublish()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestPublishDecodeError5(t *testing.T) {
	packet := []byte{
		byte(PUBLISH<<4) | 2,
		2,
		0, // topic name MSB
		1, // topic name LSB
		't',
		// < missing packet id
	}

	pkt := NewPublish()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestPublishDecodeError6(t *testing.T) {
	packet := []byte{
		byte(PUBLISH<<4) | 2,
		2,
		0, // topic name MSB
		1, // topic name LSB
		't',
		0,
		0, // < zero packet id
	}

	pkt := NewPublish()
	_, err := pkt.Decode(M4, packet)

	assert.Error(t, err)
}

func TestPublishEncode1(t *testing.T) {
	packet := []byte{
		byte(PUBLISH<<4) | 11,
		22,
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublish()
	pkt.Message.Topic = "gomqtt"
	pkt.Message.QOS = QOSAtLeastOnce
	pkt.Message.Retain = true
	pkt.Dup = true
	pkt.ID = 7
	pkt.Message.Payload = []byte("send me home")

	dst := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, dst[:n])
}

func TestPublishEncode2(t *testing.T) {
	packet := []byte{
		byte(PUBLISH << 4),
		20,
		0, // topic name MSB
		6, // topic name LSB
		'g', 'o', 'm', 'q', 't', 't',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublish()
	pkt.Message.Topic = "gomqtt"
	pkt.Message.Payload = []byte("send me home")

	dst := make([]byte, pkt.Len(M4))
	n, err := pkt.Encode(M4, dst)

	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, dst[:n])
}

func TestPublishEncodeError1(t *testing.T) {
	pkt := NewPublish()
	pkt.Message.Topic = "" // < empty topic

	dst := make([]byte, pkt.Len(M4))
	_, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
}

func TestPublishEncodeError2(t *testing.T) {
	pkt := NewPublish()
	pkt.Message.Topic = "t"
	pkt.Message.QOS = 3 // < wrong qos

	dst := make([]byte, pkt.Len(M4))
	_, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
}

func TestPublishEncodeError3(t *testing.T) {
	pkt := NewPublish()
	pkt.Message.Topic = "t"

	dst := make([]byte, 1) // < too small
	_, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
}

func TestPublishEncodeError4(t *testing.T) {
	pkt := NewPublish()
	pkt.Message.Topic = string(make([]byte, 65536)) // < too big

	dst := make([]byte, pkt.Len(M4))
	_, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
}

func TestPublishEncodeError5(t *testing.T) {
	pkt := NewPublish()
	pkt.Message.Topic = "test"
	pkt.Message.QOS = 1
	pkt.ID = 0 // < zero packet id

	dst := make([]byte, pkt.Len(M4))
	_, err := pkt.Encode(M4, dst)

	assert.Error(t, err)
}

func BenchmarkPublishEncode(b *testing.B) {
	b.ReportAllocs()

	pkt := NewPublish()
	pkt.Message.Topic = "t"
	pkt.Message.QOS = QOSAtLeastOnce
	pkt.ID = 1
	pkt.Message.Payload = []byte("p")

	buf := make([]byte, pkt.Len(M4))

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(M4, buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPublishDecode(b *testing.B) {
	b.ReportAllocs()

	packet := []byte{
		byte(PUBLISH<<4) | 2,
		6,
		0, // topic name MSB
		1, // topic name LSB
		't',
		0, // packet ID MSB
		1, // packet ID LSB
		'p',
	}

	pkt := NewPublish()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(M4, packet)
		if err != nil {
			panic(err)
		}
	}
}
