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

func TestNakedPacketDecode(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedPacketDecode(pktBytes, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n)
}

func TestNakedPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		1, // <- wrong remaining length
		0,
	}

	n, err := nakedPacketDecode(pktBytes, DISCONNECT)

	require.Error(t, err)
	require.Equal(t, 2, n)
}

func TestNakedPacketEncode(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	dst := make([]byte, nakedPacketLen())
	n, err := nakedPacketEncode(dst, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, pktBytes, dst[:n])
}

func TestNakedPacketEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedPacketDecode(pktBytes, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n)

	dst := make([]byte, nakedPacketLen())
	n2, err := nakedPacketEncode(dst, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n2)
	require.Equal(t, pktBytes, dst[:n2])

	n3, err := nakedPacketDecode(dst, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n3)
}

func BenchmarkNakedPacketEncode(b *testing.B) {
	buf := make([]byte, nakedPacketLen())

	for i := 0; i < b.N; i++ {
		_, err := nakedPacketEncode(buf, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkNakedPacketDecode(b *testing.B) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	for i := 0; i < b.N; i++ {
		_, err := nakedPacketDecode(pktBytes, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func testNakedPacketImplementation(t *testing.T, _t Type) {
	pkt, err := _t.New()
	require.NoError(t, err)
	require.Equal(t, _t, pkt.Type())
	require.NotEmpty(t, pkt.String())

	buf := make([]byte, pkt.Len())
	n, err := pkt.Encode(buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	n, err = pkt.Decode(buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)
}

func TestDisconnectImplementation(t *testing.T) {
	testNakedPacketImplementation(t, DISCONNECT)
}

func TestPingreqImplementation(t *testing.T) {
	testNakedPacketImplementation(t, PINGREQ)
}

func TestPingrespImplementation(t *testing.T) {
	testNakedPacketImplementation(t, PINGRESP)
}
