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

// Len returns the byte length of the encoded packet.
func identifiedPacketLen() int {
	return headerLen(2) + 2
}

// Decodes a identified packet.
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
		return total, 0, fmt.Errorf("%s/identifiedPacketDecode: Expected remaining length to be 2", t)
	}

	// read packet id
	packetID := binary.BigEndian.Uint16(src[total:])
	total += 2

	return total, packetID, nil
}

// Encodes a identified packet.
func identifiedPacketEncode(dst []byte, packetID uint16, mt Type) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, identifiedPacketLen(), mt)
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
	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*PubackPacket)(nil)

// NewPubackPacket creates a new PUBACK packet.
func NewPubackPacket() *PubackPacket {
	return &PubackPacket{}
}

// Type returns the packets type.
func (pm PubackPacket) Type() Type {
	return PUBACK
}

// Len returns the byte length of the encoded packet.
func (pm *PubackPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PubackPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBACK)
	pm.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubackPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pm.PacketID, PUBACK)
}

// String returns a string representation of the packet.
func (pm PubackPacket) String() string {
	return fmt.Sprintf("PUBACK: PacketID=%d", pm.PacketID)
}

// The PubcompPacket is the response to a PubrelPacket. It is the fourth and
// final packet of the QOS 2 protocol exchange.
type PubcompPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*PubcompPacket)(nil)

// NewPubcompPacket creates a new PUBCOMP packet.
func NewPubcompPacket() *PubcompPacket {
	return &PubcompPacket{}
}

// Type returns the packets type.
func (pm PubcompPacket) Type() Type {
	return PUBCOMP
}

// Len returns the byte length of the encoded packet.
func (pm *PubcompPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PubcompPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBCOMP)
	pm.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubcompPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pm.PacketID, PUBCOMP)
}

// String returns a string representation of the packet.
func (pm PubcompPacket) String() string {
	return fmt.Sprintf("PUBCOMP: PacketID=%d", pm.PacketID)
}

// A PubrecPacket is the response to a PublishPacket with QOS 2. It is the second
// packet of the QOS 2 protocol exchange.
type PubrecPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*PubrecPacket)(nil)

// NewPubrecPacket creates a new PUBREC packet.
func NewPubrecPacket() *PubrecPacket {
	return &PubrecPacket{}
}

// Type returns the packets type.
func (pm PubrecPacket) Type() Type {
	return PUBREC
}

// Len returns the byte length of the encoded packet.
func (pm *PubrecPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PubrecPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBREC)
	pm.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubrecPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pm.PacketID, PUBREC)
}

// String returns a string representation of the packet.
func (pm PubrecPacket) String() string {
	return fmt.Sprintf("PUBREC: PacketID=%d", pm.PacketID)
}

// A PubrelPacket is the response to a PubrecPacket. It is the third packet of the
// QOS 2 protocol exchange.
type PubrelPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*PubrelPacket)(nil)

// NewPubrelPacket creates a new PUBREL packet.
func NewPubrelPacket() *PubrelPacket {
	return &PubrelPacket{}
}

// Type returns the packets type.
func (pm PubrelPacket) Type() Type {
	return PUBREL
}

// Len returns the byte length of the encoded packet.
func (pm *PubrelPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PubrelPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBREL)
	pm.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubrelPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pm.PacketID, PUBREL)
}

// String returns a string representation of the packet.
func (pm PubrelPacket) String() string {
	return fmt.Sprintf("PUBREL: PacketID=%d", pm.PacketID)
}

// The UnsubackPacket is sent by the Server to the Client to confirm receipt of an
// UnsubscribePacket.
type UnsubackPacket struct {
	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*UnsubackPacket)(nil)

// NewUnsubackPacket creates a new UNSUBACK packet.
func NewUnsubackPacket() *UnsubackPacket {
	return &UnsubackPacket{}
}

// Type returns the packets type.
func (um UnsubackPacket) Type() Type {
	return UNSUBACK
}

// Len returns the byte length of the encoded packet.
func (um *UnsubackPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (um *UnsubackPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, UNSUBACK)
	um.PacketID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (um *UnsubackPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, um.PacketID, UNSUBACK)
}

// String returns a string representation of the packet.
func (um UnsubackPacket) String() string {
	return fmt.Sprintf("UNSUBACK: PacketID=%d", um.PacketID)
}
