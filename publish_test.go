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

	"github.com/stretchr/testify/require"
)

func TestPublishInterface(t *testing.T) {
	pkt := NewPublishPacket()

	require.Equal(t, pkt.Type(), PUBLISH)
	require.NotNil(t, pkt.String())
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

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, 7, int(pkt.PacketID))
	require.Equal(t, []byte("surgemq"), pkt.Topic)
	require.Equal(t, []byte("send me home"), pkt.Payload)
	require.Equal(t, 1, int(pkt.QOS))
	require.Equal(t, true, pkt.Retain)
	require.Equal(t, true, pkt.Dup)
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

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, 0, int(pkt.PacketID))
	require.Equal(t, []byte("surgemq"), pkt.Topic)
	require.Equal(t, []byte("send me home"), pkt.Payload)
	require.Equal(t, 0, int(pkt.QOS))
	require.Equal(t, false, pkt.Retain)
	require.Equal(t, false, pkt.Dup)
}

func TestPublishPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		2, // <- too much
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestPublishPacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH<<4) | 6, // <- wrong qos
		0,
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestPublishPacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(PUBLISH << 4),
		0,
		// <- missing topic stuff
	}

	pkt := NewPublishPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
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

	require.Error(t, err)
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

	require.Error(t, err)
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
	pkt.Topic = []byte("surgemq")
	pkt.QOS = QOSAtLeastOnce
	pkt.Retain = true
	pkt.Dup = true
	pkt.PacketID = 7
	pkt.Payload = []byte("send me home")

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, pktBytes, dst[:n])
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
	pkt.Topic = []byte("surgemq")
	pkt.Payload = []byte("send me home")

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, pktBytes, dst[:n])
}

func TestPublishPacketEncodeError1(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Topic = []byte("") // <- empty topic

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	require.Error(t, err)
}

func TestPublishPacketEncodeError2(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Topic = []byte("t")
	pkt.QOS = 3 // <- wrong qos

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	require.Error(t, err)
}

func TestPublishPacketEncodeError3(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Topic = []byte("t")

	dst := make([]byte, 1) // <- too small
	_, err := pkt.Encode(dst)

	require.Error(t, err)
}

func TestPublishPacketEncodeError4(t *testing.T) {
	pkt := NewPublishPacket()
	pkt.Topic = make([]byte, 65536) // <- too big

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	require.Error(t, err)
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

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)

	dst := make([]byte, pkt.Len())
	n2, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n2)
	require.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n3)
}

func BenchmarkPublishEncode(b *testing.B) {
	pkt := NewPublishPacket()
	pkt.Topic = []byte("t")
	pkt.QOS = QOSAtLeastOnce
	pkt.PacketID = 1
	pkt.Payload = []byte("p")

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
