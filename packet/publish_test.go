// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishInterface(t *testing.T) {
	pkt := NewPublishPacket()

	assert.Equal(t, pkt.Type(), PUBLISH)
	assert.Equal(t, "<PublishPacket PacketID=0 Message=<Message Topic=\"\" QOS=0 Retain=false Payload=[]> Dup=false>", pkt.String())
}

func TestPublishPacketDecode1(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 11,
		23,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublishPacket()
	n, err := pkt.Decode(pktBytes)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)
	assert.Equal(t, uint16(7), pkt.PacketID)
	assert.Equal(t, "surgemq", pkt.Message.Topic)
	assert.Equal(t, []byte("send me home"), pkt.Message.Payload)
	assert.Equal(t, uint8(1), pkt.Message.QOS)
	assert.Equal(t, true, pkt.Message.Retain)
	assert.Equal(t, true, pkt.Dup)
}

func TestPublishPacketDecode2(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublishPacket()
	n, err := pkt.Decode(pktBytes)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)
	assert.Equal(t, uint16(0), pkt.PacketID)
	assert.Equal(t, "surgemq", pkt.Message.Topic)
	assert.Equal(t, []byte("send me home"), pkt.Message.Payload)
	assert.Equal(t, uint8(0), pkt.Message.QOS)
	assert.Equal(t, false, pkt.Message.Retain)
	assert.Equal(t, false, pkt.Dup)
}

func TestPublishPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		2, // <- too much
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestPublishPacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 6, // <- wrong qos
		0,
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestPublishPacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		0,
		// <- missing topic stuff
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestPublishPacketDecodeError4(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		2,
		0, // topic name MSB
		1, // topic name LSB
		// <- missing topic string
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestPublishPacketDecodeError5(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 2,
		2,
		0, // topic name MSB
		1, // topic name LSB
		't',
		// <- missing packet id
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestPublishPacketDecodeError6(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 2,
		2,
		0, // topic name MSB
		1, // topic name LSB
		't',
		0,
		0, // <- zero packet id
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestPublishPacketEncode1(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 11,
		23,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublishPacket()
	pkt.Message.Topic = "surgemq"
	pkt.Message.QOS = QOSAtLeastOnce
	pkt.Message.Retain = true
	pkt.Dup = true
	pkt.PacketID = 7
	pkt.Message.Payload = []byte("send me home")

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)
	assert.Equal(t, pktBytes, dst[:n])
}

func TestPublishPacketEncode2(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublishPacket()
	pkt.Message.Topic = "surgemq"
	pkt.Message.Payload = []byte("send me home")

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)
	assert.Equal(t, pktBytes, dst[:n])
}

func TestPublishPacketEncodeError1(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Message.Topic = "" // <- empty topic

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestPublishPacketEncodeError2(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Message.Topic = "t"
	pkt.Message.QOS = 3 // <- wrong qos

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestPublishPacketEncodeError3(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Message.Topic = "t"

	dst := make([]byte, 1) // <- too small
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestPublishPacketEncodeError4(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Message.Topic = string(make([]byte, 65536)) // <- too big

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestPublishPacketEncodeError5(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Message.Topic = "test"
	pkt.Message.QOS = 1
	pkt.PacketID = 0 // <- zero packet id

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	assert.Error(t, err)
}

func TestPublishEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	pkt := NewPublishPacket()
	n, err := pkt.Decode(pktBytes)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)

	dst := make([]byte, pkt.Len())
	n2, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n2)
	assert.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n3)
}

func BenchmarkPublishEncode(b *testing.B) {
	pkt := NewPublishPacket()
	pkt.Message.Topic = "t"
	pkt.Message.QOS = QOSAtLeastOnce
	pkt.PacketID = 1
	pkt.Message.Payload = []byte("p")

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPublishDecode(b *testing.B) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 2,
		6,
		0, // topic name MSB
		1, // topic name LSB
		't',
		0, // packet ID MSB
		1, // packet ID LSB
		'p',
	}

	pkt := NewPublishPacket()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}
