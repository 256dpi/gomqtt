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

func TestIdentifiedPacketDecode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB
		7, // packet ID LSB
	}

	n, pid, err := identifiedPacketDecode(msgBytes, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, 7, int(pid))
}

func TestIdentifiedPacketDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		1, // <- wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
	}

	n, pid, err := identifiedPacketDecode(msgBytes, PUBACK)

	require.Error(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, 0, int(pid))
}

func TestIdentifiedPacketDecodeError2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		7, // packet ID LSB
		// <- insufficient bytes
	}

	n, pid, err := identifiedPacketDecode(msgBytes, PUBACK)

	require.Error(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, 0, int(pid))
}

func TestIdentifiedPacketEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB
		7, // packet ID LSB
	}

	dst := make([]byte, identifiedPacketLen())
	n, err := identifiedPacketEncode(dst, 7, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestIdentifiedPacketEncodeError1(t *testing.T) {
	dst := make([]byte, 3) // <- insufficient buffer
	n, err := identifiedPacketEncode(dst, 7, PUBACK)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestIdentifiedPacketEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB
		7, // packet ID LSB
	}

	msg := &PubackPacket{}
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, 4, n)

	dst := make([]byte, 100)
	n2, err := identifiedPacketEncode(dst, 7, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, pid, err := identifiedPacketDecode(msgBytes, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n3)
	require.Equal(t, 7, int(pid))
}

func BenchmarkIdentifiedPacketEncode(b *testing.B) {
	msg := &PubackPacket{}
	msg.PacketID = 1

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkIdentifiedPacketDecode(b *testing.B) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB
		1, // packet ID LSB
	}

	msg := &PubackPacket{}

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}

func testIdentifiedPacketImplementation(t *testing.T, _t Type) {
	msg, err := _t.New()
	require.NoError(t, err)
	require.Equal(t, _t, msg.Type())
	require.NotEmpty(t, msg.String())

	buf := make([]byte, msg.Len())
	n, err := msg.Encode(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)

	n, err = msg.Decode(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
}

func TestPubackImplementation(t *testing.T) {
	testIdentifiedPacketImplementation(t, PUBACK)
}

func TestPubcompImplementation(t *testing.T) {
	testIdentifiedPacketImplementation(t, PUBCOMP)
}

func TestPubrecImplementation(t *testing.T) {
	testIdentifiedPacketImplementation(t, PUBREC)
}

func TestPubrelImplementation(t *testing.T) {
	testIdentifiedPacketImplementation(t, PUBREL)
}

func TestUnsubackImplementation(t *testing.T) {
	testIdentifiedPacketImplementation(t, UNSUBACK)
}
