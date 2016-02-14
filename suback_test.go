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

func TestSubackInterface(t *testing.T) {
	pkt := NewSubackPacket()
	pkt.ReturnCodes = []byte{0, 1}

	assert.Equal(t, pkt.Type(), SUBACK)
	assert.Equal(t, "<SubackPacket PacketID=0 ReturnCodes=[0, 1]>", pkt.String())
}

func TestSubackPacketDecode(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	pkt := NewSubackPacket()
	n, err := pkt.Decode(pktBytes)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)
	assert.Equal(t, 4, len(pkt.ReturnCodes))
}

func TestSubackPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		1, // <- wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
		0, // return code 1
	}

	pkt := NewSubackPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestSubackPacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x81, // <- wrong return code
	}

	pkt := NewSubackPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestSubackPacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		1, // <- wrong remaining length
		0, // packet ID MSB
	}

	pkt := NewSubackPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestSubackPacketDecodeError4(t *testing.T) {
	pktBytes := []byte{
		byte(PUBCOMP << 4), // <- wrong packet type
		3,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // return code 1
	}

	pkt := NewSubackPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestSubackPacketDecodeError5(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		3,
		0, // packet ID MSB
		0, // packet ID LSB <- zero packet id
		0,
	}

	pkt := NewSubackPacket()
	_, err := pkt.Decode(pktBytes)

	assert.Error(t, err)
}

func TestSubackPacketEncode(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	pkt := NewSubackPacket()
	pkt.PacketID = 7
	pkt.ReturnCodes = []byte{0, 1, 2, 0x80}

	dst := make([]byte, 10)
	n, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)
	assert.Equal(t, pktBytes, dst[:n])
}

func TestSubackPacketEncodeError1(t *testing.T) {
	pkt := NewSubackPacket()
	pkt.PacketID = 7
	pkt.ReturnCodes = []byte{0x81}

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSubackPacketEncodeError2(t *testing.T) {
	pkt := NewSubackPacket()
	pkt.PacketID = 7
	pkt.ReturnCodes = []byte{0x80}

	dst := make([]byte, pkt.Len()-1)
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSubackPacketEncodeError3(t *testing.T) {
	pkt := NewSubackPacket()
	pkt.PacketID = 0 // <- zero packet id
	pkt.ReturnCodes = []byte{0x80}

	dst := make([]byte, pkt.Len()-1)
	n, err := pkt.Encode(dst)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSubackEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		6,
		0,    // packet ID MSB
		7,    // packet ID LSB
		0,    // return code 1
		1,    // return code 2
		2,    // return code 3
		0x80, // return code 4
	}

	pkt := NewSubackPacket()
	n, err := pkt.Decode(pktBytes)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n)

	dst := make([]byte, 100)
	n2, err := pkt.Encode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n2)
	assert.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	assert.NoError(t, err)
	assert.Equal(t, len(pktBytes), n3)
}

func BenchmarkSubackEncode(b *testing.B) {
	pkt := NewSubackPacket()
	pkt.PacketID = 1
	pkt.ReturnCodes = []byte{0}

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSubackDecode(b *testing.B) {
	pktBytes := []byte{
		byte(SUBACK << 4),
		3,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // return code 1
	}

	pkt := NewSubackPacket()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}
