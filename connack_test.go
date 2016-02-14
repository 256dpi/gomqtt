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

func TestConnackReturnCodes(t *testing.T) {
	assert.Equal(t, ConnectionAccepted.Error(), ConnackCode(0).Error())
	assert.Equal(t, ErrInvalidProtocolVersion.Error(), ConnackCode(1).Error())
	assert.Equal(t, ErrIdentifierRejected.Error(), ConnackCode(2).Error())
	assert.Equal(t, ErrServerUnavailable.Error(), ConnackCode(3).Error())
	assert.Equal(t, ErrBadUsernameOrPassword.Error(), ConnackCode(4).Error())
	assert.Equal(t, ErrNotAuthorized.Error(), ConnackCode(5).Error())
	assert.Equal(t, "Unknown error", ConnackCode(6).Error())
}

func TestConnackInterface(t *testing.T) {
	pkt := NewConnackPacket()

	assert.Equal(t, pkt.Type(), CONNACK)
	assert.Equal(t, "<ConnackPacket SessionPresent=false ReturnCode=0>", pkt.String())
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

	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.False(t, pkt.SessionPresent)
	assert.Equal(t, ConnectionAccepted, pkt.ReturnCode)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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

	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, pktBytes, dst[:n])
}

func TestConnackPacketEncodeError1(t *testing.T) {
	pkt := NewConnackPacket()

	dst := make([]byte, 3) // <- wrong buffer size
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestConnackPacketEncodeError2(t *testing.T) {
	pkt := NewConnackPacket()
	pkt.ReturnCode = 11 // <- wrong return code

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 3, n)
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

	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	dst := make([]byte, pkt.Len())
	n2, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, 4, n2)
	assert.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	assert.NoError(t, err)
	assert.Equal(t, 4, n3)
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
