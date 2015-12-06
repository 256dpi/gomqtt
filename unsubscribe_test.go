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

func TestUnsubscribeInterface(t *testing.T) {
	pkt := NewUnsubscribePacket()
	pkt.Topics = [][]byte{[]byte("hello")}

	require.Equal(t, pkt.Type(), UNSUBSCRIBE)
	require.NotNil(t, pkt.String())
}

func TestUnsubscribePacketDecode(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		33,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	pkt := NewUnsubscribePacket()
	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, 3, len(pkt.Topics))
	require.Equal(t, []byte("surgemq"), pkt.Topics[0])
	require.Equal(t, []byte("/a/b/#/c"), pkt.Topics[1])
	require.Equal(t, []byte("/a/b/#/cdd"), pkt.Topics[2])
}

func TestUnsubscribePacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		2,
		0, // packet ID MSB
		7, // packet ID LSB
		// empty topic list
	}

	pkt := NewUnsubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestUnsubscribePacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		6, // <- wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
	}

	pkt := NewUnsubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestUnsubscribePacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		0,
		// missing packet id
	}

	pkt := NewUnsubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestUnsubscribePacketDecodeError4(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		11,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		9, // topic name LSB <- wrong size
		's', 'u', 'r', 'g', 'e', 'm', 'q',
	}

	pkt := NewUnsubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestUnsubscribePacketEncode(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		33,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	pkt := NewUnsubscribePacket()
	pkt.PacketID = 7
	pkt.Topics = [][]byte{
		[]byte("surgemq"),
		[]byte("/a/b/#/c"),
		[]byte("/a/b/#/cdd"),
	}

	dst := make([]byte, 100)
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, pktBytes, dst[:n])
}

func TestUnsubscribePacketEncodeError1(t *testing.T) {
	pkt := NewUnsubscribePacket()
	pkt.PacketID = 7
	pkt.Topics = [][]byte{
		[]byte("surgemq"),
	}

	dst := make([]byte, 1) // <- too small
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestUnsubscribePacketEncodeError2(t *testing.T) {
	pkt := NewUnsubscribePacket()
	pkt.PacketID = 7
	pkt.Topics = [][]byte{
		make([]byte, 65536),
	}

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 6, n)
}

// test to ensure encoding and decoding are the same
// decode, encode, and decode again
func TestUnsubscribeEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		33,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	pkt := NewUnsubscribePacket()
	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)

	dst := make([]byte, 100)
	n2, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n2)
	require.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n3)
}

func BenchmarkUnsubscribeEncode(b *testing.B) {
	pkt := NewUnsubscribePacket()
	pkt.PacketID = 1
	pkt.Topics = [][]byte{
		[]byte("t"),
	}

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkUnsubscribeDecode(b *testing.B) {
	pktBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		't',
	}

	pkt := NewUnsubscribePacket()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}
