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

func TestSubackInterface(t *testing.T) {
	msg := NewSubackPacket()

	require.Equal(t, msg.Type(), SUBACK)
	require.NotNil(t, msg.String())
}

func TestSubackPacketDecode(t *testing.T) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	msg := NewSubackPacket()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, 4, len(msg.ReturnCodes))
}

func TestSubackPacketDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		1, // <- wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
		0, // return code 1
	}

	msg := NewSubackPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestSubackPacketDecodeError2(t *testing.T) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x81, // <- wrong return code
	}

	msg := NewSubackPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestSubackPacketDecodeError3(t *testing.T) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		1, // <- wrong remaining length
		0, // packet ID MSB
	}

	msg := NewSubackPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestSubackPacketDecodeError4(t *testing.T) {
	msgBytes := []byte{
		byte(PUBCOMP << 4), // <- wrong packet type
		3,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // return code 1
	}

	msg := NewSubackPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestSubackPacketEncode(t *testing.T) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	msg := NewSubackPacket()
	msg.PacketID = 7
	msg.ReturnCodes = []byte{0, 1, 2, 0x80}

	dst := make([]byte, 10)
	n, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestSubackPacketEncodeError1(t *testing.T) {
	msg := NewSubackPacket()
	msg.PacketID = 7
	msg.ReturnCodes = []byte{0x81}

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestSubackPacketEncodeError2(t *testing.T) {
	msg := NewSubackPacket()
	msg.PacketID = 7
	msg.ReturnCodes = []byte{0x80}

	dst := make([]byte, msg.Len()-1)
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestSubackEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	msg := NewSubackPacket()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)

	dst := make([]byte, 100)
	n2, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, err := msg.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n3)
}

func BenchmarkSubackEncode(b *testing.B) {
	msg := NewSubackPacket()
	msg.PacketID = 1
	msg.ReturnCodes = []byte{0}

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSubackDecode(b *testing.B) {
	msgBytes := []byte{
		byte(SUBACK << 4),
		3,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // return code 1
	}

	msg := NewSubackPacket()

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}
