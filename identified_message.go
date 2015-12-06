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

package message

import (
	"encoding/binary"
	"fmt"
)

// Len returns the byte length of the message.
func identifiedMessageLen() int {
	return headerLen(2) + 2
}

// Decodes a identified message.
func identifiedMessageDecode(src []byte, mt MessageType) (int, uint16, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, mt)
	total += hl
	if err != nil {
		return total, 0, err
	}

	// check remaining length
	if rl != 2 {
		return total, 0, fmt.Errorf("%s/identifiedMessageDecode: Expected remaining length to be 2", mt)
	}

	// read packet id
	packetId := binary.BigEndian.Uint16(src[total:])
	total += 2

	return total, packetId, nil
}

// Encodes a identified message.
func identifiedMessageEncode(dst []byte, packetId uint16, mt MessageType) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, identifiedMessageLen(), mt)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], packetId)
	total += 2

	return total, nil
}

// A PUBACK Packet is the response to a PUBLISH Packet with QoS level 1.
type PubackMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubackMessage)(nil)

// NewPubackMessage creates a new PUBACK message.
func NewPubackMessage() *PubackMessage {
	return &PubackMessage{}
}

// Type return the messages message type.
func (pm PubackMessage) Type() MessageType {
	return PUBACK
}

// Len returns the byte length of the message.
func (pm *PubackMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (pm *PubackMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBACK)
	pm.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubackMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, pm.PacketId, PUBACK)
}

// String returns a string representation of the message.
func (pm PubackMessage) String() string {
	return fmt.Sprintf("PUBACK: PacketId=%d", pm.PacketId)
}

// The PUBCOMP Packet is the response to a PUBREL Packet. It is the fourth and
// final packet of the QoS 2 protocol exchange.
type PubcompMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubcompMessage)(nil)

// NewPubcompMessage creates a new PUBCOMP message.
func NewPubcompMessage() *PubcompMessage {
	return &PubcompMessage{}
}

// Type return the messages message type.
func (pm PubcompMessage) Type() MessageType {
	return PUBCOMP
}

// Len returns the byte length of the message.
func (pm *PubcompMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (pm *PubcompMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBCOMP)
	pm.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubcompMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, pm.PacketId, PUBCOMP)
}

// String returns a string representation of the message.
func (pm PubcompMessage) String() string {
	return fmt.Sprintf("PUBCOMP: PacketId=%d", pm.PacketId)
}

// A PUBREC Packet is the response to a PUBLISH Packet with QoS 2. It is the second
// packet of the QoS 2 protocol exchange.
type PubrecMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubrecMessage)(nil)

// NewPubrecMessage creates a new PUBREC message.
func NewPubrecMessage() *PubrecMessage {
	return &PubrecMessage{}
}

// Type return the messages message type.
func (pm PubrecMessage) Type() MessageType {
	return PUBREC
}

// Len returns the byte length of the message.
func (pm *PubrecMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (pm *PubrecMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBREC)
	pm.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubrecMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, pm.PacketId, PUBREC)
}

// String returns a string representation of the message.
func (pm PubrecMessage) String() string {
	return fmt.Sprintf("PUBREC: PacketId=%d", pm.PacketId)
}

// A PUBREL Packet is the response to a PUBREC Packet. It is the third packet of the
// QoS 2 protocol exchange.
type PubrelMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubrelMessage)(nil)

// NewPubrelMessage creates a new PUBREL message.
func NewPubrelMessage() *PubrelMessage {
	return &PubrelMessage{}
}

// Type return the messages message type.
func (pm PubrelMessage) Type() MessageType {
	return PUBREL
}

// Len returns the byte length of the message.
func (pm *PubrelMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (pm *PubrelMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBREL)
	pm.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PubrelMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, pm.PacketId, PUBREL)
}

// String returns a string representation of the message.
func (pm PubrelMessage) String() string {
	return fmt.Sprintf("PUBREL: PacketId=%d", pm.PacketId)
}

// The UNSUBACK Packet is sent by the Server to the Client to confirm receipt of an
// UNSUBSCRIBE Packet.
type UnsubackMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*UnsubackMessage)(nil)

// NewUnsubackMessage creates a new UNSUBACK message.
func NewUnsubackMessage() *UnsubackMessage {
	return &UnsubackMessage{}
}

// Type return the messages message type.
func (um UnsubackMessage) Type() MessageType {
	return UNSUBACK
}

// Len returns the byte length of the message.
func (um *UnsubackMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (um *UnsubackMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, UNSUBACK)
	um.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (um *UnsubackMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, um.PacketId, UNSUBACK)
}

// String returns a string representation of the message.
func (um UnsubackMessage) String() string {
	return fmt.Sprintf("UNSUBACK: PacketId=%d", um.PacketId)
}
