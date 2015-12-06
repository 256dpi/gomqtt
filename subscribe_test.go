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

func TestSubscribeInterface(t *testing.T) {
	pkt := NewSubscribePacket()
	pkt.Subscriptions = []Subscription{
		{Topic: []byte("hello"), QOS: QOSAtMostOnce},
	}

	require.Equal(t, pkt.Type(), SUBSCRIBE)
	require.NotNil(t, pkt.String())
}

func TestSubscribePacketDecode(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		36,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
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

	pkt := NewSubscribePacket()
	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, 3, len(pkt.Subscriptions))
	require.Equal(t, []byte("surgemq"), pkt.Subscriptions[0].Topic)
	require.Equal(t, 0, int(pkt.Subscriptions[0].QOS))
	require.Equal(t, []byte("/a/b/#/c"), pkt.Subscriptions[1].Topic)
	require.Equal(t, 1, int(pkt.Subscriptions[1].QOS))
	require.Equal(t, []byte("/a/b/#/cdd"), pkt.Subscriptions[2].Topic)
	require.Equal(t, 2, int(pkt.Subscriptions[2].QOS))
}

func TestSubscribePacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		9, // <- too much
	}

	pkt := NewSubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestSubscribePacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		0,
		// <- missing packet id
	}

	pkt := NewSubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestSubscribePacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		2,
		0, // packet ID MSB
		7, // packet ID LSB
		// <- missing subscription
	}

	pkt := NewSubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestSubscribePacketDecodeError4(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		2, // topic name LSB <- wrong size
		's',
	}

	pkt := NewSubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestSubscribePacketDecodeError5(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		's',
		// <- missing qos
	}

	pkt := NewSubscribePacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestSubscribePacketEncode(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		36,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
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

	pkt := NewSubscribePacket()
	pkt.PacketID = 7
	pkt.Subscriptions = []Subscription{
		{[]byte("surgemq"), 0},
		{[]byte("/a/b/#/c"), 1},
		{[]byte("/a/b/#/cdd"), 2},
	}

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, pktBytes, dst)
}

func TestSubscribePacketEncodeError1(t *testing.T) {
	pkt := NewSubscribePacket()

	dst := make([]byte, 1) // <- too small
	_, err := pkt.Encode(dst)

	require.Error(t, err)
}

func TestSubscribePacketEncodeError2(t *testing.T) {
	pkt := NewSubscribePacket()
	pkt.Subscriptions = []Subscription{
		{make([]byte, 65536), 0}, // too big
	}

	dst := make([]byte, pkt.Len())
	_, err := pkt.Encode(dst)

	require.Error(t, err)
}

func TestSubscribeEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		36,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
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

	pkt := NewSubscribePacket()
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

func BenchmarkSubscribeEncode(b *testing.B) {
	pkt := NewSubscribePacket()
	pkt.PacketID = 7
	pkt.Subscriptions = []Subscription{
		{[]byte("t"), 0},
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
	pktBytes := []byte{
		byte(SUBSCRIBE<<4) | 2,
		6,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		't',
		0, // QOS
	}

	pkt := NewSubscribePacket()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}
