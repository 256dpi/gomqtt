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

func TestConnackReturnCodes(t *testing.T) {
	require.Equal(t, ErrInvalidProtocolVersion.Error(), ConnackCode(1).Error())
	require.Equal(t, ErrIdentifierRejected.Error(), ConnackCode(2).Error())
	require.Equal(t, ErrServerUnavailable.Error(), ConnackCode(3).Error())
	require.Equal(t, ErrBadUsernameOrPassword.Error(), ConnackCode(4).Error())
	require.Equal(t, ErrNotAuthorized.Error(), ConnackCode(5).Error())
	require.Equal(t, "Unknown error", ConnackCode(6).Error())
}

func TestConnackInterface(t *testing.T) {
	pkt := NewConnackPacket()

	require.Equal(t, pkt.Type(), CONNACK)
	require.NotNil(t, pkt.String())
}

func TestConnackPacketDecode(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnackPacket()

	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.False(t, pkt.SessionPresent)
	require.Equal(t, ConnectionAccepted, pkt.ReturnCode)
}

func TestConnackPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		3, // <- wrong size
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnackPacket()

	_, err := pkt.Decode(pktBytes)
	require.Error(t, err)
}

func TestConnackPacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		// <- wrong packet size
	}

	pkt := NewConnackPacket()

	_, err := pkt.Decode(pktBytes)
	require.Error(t, err)
}

func TestConnackPacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		64, // <- wrong value
		0,  // connection accepted
	}

	pkt := NewConnackPacket()

	_, err := pkt.Decode(pktBytes)
	require.Error(t, err)
}

func TestConnackPacketDecodeError4(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		0,
		6, // <- wrong code
	}

	pkt := NewConnackPacket()

	_, err := pkt.Decode(pktBytes)
	require.Error(t, err)
}

func TestConnackPacketDecodeError5(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		1, // <- wrong remaining length
		0,
		6,
	}

	pkt := NewConnackPacket()

	_, err := pkt.Decode(pktBytes)
	require.Error(t, err)
}

func TestConnackPacketEncode(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		1, // session present
		0, // connection accepted
	}

	pkt := NewConnackPacket()
	pkt.ReturnCode = ConnectionAccepted
	pkt.SessionPresent = true

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, pktBytes, dst[:n])
}

func TestConnackPacketEncodeError1(t *testing.T) {
	pkt := NewConnackPacket()

	dst := make([]byte, 3) // <- wrong buffer size
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestConnackPacketEncodeError2(t *testing.T) {
	pkt := NewConnackPacket()
	pkt.ReturnCode = 11 // <- wrong return code

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 3, n)
}

func TestConnackEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnackPacket()
	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, 4, n)

	dst := make([]byte, pkt.Len())
	n2, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, 4, n2)
	require.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, 4, n3)
}

func BenchmarkConnackEncode(b *testing.B) {
	pkt := NewConnackPacket()
	pkt.ReturnCode = ConnectionAccepted
	pkt.SessionPresent = true

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkConnackDecode(b *testing.B) {
	pktBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	pkt := NewConnackPacket()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}
