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
	"encoding/binary"
	"fmt"
)

// Returns the byte length of an identified packet.
func identifiedPacketLen() int {
	return headerLen(2) + 2
}

// Decodes an identified packet.
func identifiedPacketDecode(src []byte, t Type) (int, uint16, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, t)
	total += hl
	if err != nil {
		return total, 0, err
	}

	// check remaining length
	if rl != 2 {
		return total, 0, fmt.Errorf("[%s] expected remaining length to be 2", t)
	}

	// read packet id
	packetID := binary.BigEndian.Uint16(src[total:])
	total += 2

	// check packet id
	if packetID == 0 {
		return total, 0, fmt.Errorf("[%s] packet id must be grater than zero", t)
	}

	return total, packetID, nil
}

// Encodes an identified packet.
func identifiedPacketEncode(dst []byte, packetID uint16, t Type) (int, error) {
	total := 0

	// check packet id
	if packetID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", t)
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, identifiedPacketLen(), t)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], packetID)
	total += 2

	return total, nil
}

// A PubackPacket is the response to a PublishPacket with QOS level 1.
type PubackPacket struct {
	// The packet identifier.
	PacketID uint16
}

// NewPubackPacket creates a new PubackPacket.
func NewPubackPacket() *PubackPacket {
	return &PubackPacket{}
}

// Type returns the packets type.
func (pp *PubackPacket) Type() Type {
	return PUBACK
}

// Len returns the byte length of the encoded packet.
func (pp *PubackPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubackPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBACK)
	pp.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubackPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.PacketID, PUBACK)
}

// String returns a string representation of the packet.
func (pp *PubackPacket) String() string {
	return fmt.Sprintf("<PubackPacket PacketID=%d>", pp.PacketID)
}

// A PubcompPacket is the response to a PubrelPacket. It is the fourth and
// final packet of the QOS 2 protocol exchange.
type PubcompPacket struct {
	// The packet identifier.
	PacketID uint16
}

var _ Packet = (*PubcompPacket)(nil)

// NewPubcompPacket creates a new PubcompPacket.
func NewPubcompPacket() *PubcompPacket {
	return &PubcompPacket{}
}

// Type returns the packets type.
func (pp *PubcompPacket) Type() Type {
	return PUBCOMP
}

// Len returns the byte length of the encoded packet.
func (pp *PubcompPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubcompPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBCOMP)
	pp.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubcompPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.PacketID, PUBCOMP)
}

// String returns a string representation of the packet.
func (pp *PubcompPacket) String() string {
	return fmt.Sprintf("<PubcompPacket PacketID=%d>", pp.PacketID)
}

// A PubrecPacket is the response to a PublishPacket with QOS 2. It is the
// second packet of the QOS 2 protocol exchange.
type PubrecPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

// NewPubrecPacket creates a new PubrecPacket.
func NewPubrecPacket() *PubrecPacket {
	return &PubrecPacket{}
}

// Type returns the packets type.
func (pp *PubrecPacket) Type() Type {
	return PUBREC
}

// Len returns the byte length of the encoded packet.
func (pp *PubrecPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubrecPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBREC)
	pp.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubrecPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.PacketID, PUBREC)
}

// String returns a string representation of the packet.
func (pp *PubrecPacket) String() string {
	return fmt.Sprintf("<PubrecPacket PacketID=%d>", pp.PacketID)
}

// A PubrelPacket is the response to a PubrecPacket. It is the third packet of
// the QOS 2 protocol exchange.
type PubrelPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*PubrelPacket)(nil)

// NewPubrelPacket creates a new PubrelPacket.
func NewPubrelPacket() *PubrelPacket {
	return &PubrelPacket{}
}

// Type returns the packets type.
func (pp *PubrelPacket) Type() Type {
	return PUBREL
}

// Len returns the byte length of the encoded packet.
func (pp *PubrelPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubrelPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBREL)
	pp.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubrelPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.PacketID, PUBREL)
}

// String returns a string representation of the packet.
func (pp *PubrelPacket) String() string {
	return fmt.Sprintf("<PubrelPacket PacketID=%d>", pp.PacketID)
}

// An UnsubackPacket is sent by the server to the client to confirm receipt of
// an UnsubscribePacket.
type UnsubackPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

// NewUnsubackPacket creates a new UnsubackPacket.
func NewUnsubackPacket() *UnsubackPacket {
	return &UnsubackPacket{}
}

// Type returns the packets type.
func (up *UnsubackPacket) Type() Type {
	return UNSUBACK
}

// Len returns the byte length of the encoded packet.
func (up *UnsubackPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (up *UnsubackPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, UNSUBACK)
	up.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (up *UnsubackPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, up.PacketID, UNSUBACK)
}

// String returns a string representation of the packet.
func (up *UnsubackPacket) String() string {
	return fmt.Sprintf("<UnsubackPacket PacketID=%d>", up.PacketID)
}
