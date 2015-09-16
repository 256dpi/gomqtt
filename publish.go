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

// A PUBLISH Control Packet is sent from a Client to a Server or from Server to a Client
// to transport an Application Message.
type PublishMessage struct {
	header

	// The Topic of the message.
	Topic []byte

	// The Payload of the message.
	Payload []byte

	// The QoS indicates the level of assurance for delivery of a message.
	QoS byte

	// If the RETAIN flag is set to true, in a PUBLISH Packet sent by a Client to a
	// Server, the Server MUST store the Application Message and its QoS, so that it can be
	// delivered to future subscribers whose subscriptions match its topic name.
	Retain bool

	// If the DUP flag is set to false, it indicates that this is the first occasion that the
	// Client or Server has attempted to send this MQTT PUBLISH Packet. If the DUP flag is
	// set to true, it indicates that this might be re-delivery of an earlier attempt to send
	// the Packet.
	Dup bool

	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PublishMessage)(nil)

// NewPublishMessage creates a new PUBLISH message.
func NewPublishMessage() *PublishMessage {
	msg := &PublishMessage{}
	msg.messageType = PUBLISH
	return msg
}

func (this PublishMessage) Type() MessageType {
	return PUBLISH
}

// String returns a string representation of the message.
func (this PublishMessage) String() string {
	return fmt.Sprintf("PUBLISH: Topic=%q PacketId=%d QoS=%d Retained=%t Dup=%t Payload=%v",
		this.Topic, this.PacketId, this.QoS, this.Retain, this.Dup, this.Payload)
}

// Len returns the byte length of the message.
func (this *PublishMessage) Len() int {
	ml := this.msglen()
	return this.header.len(ml) + ml
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *PublishMessage) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, flags, rl, err := this.header.decode(src[total:])
	total += hl
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("PUBLISH/Decode: Insufficient buffer size. Expecting %d, got %d.", total+2, len(src))
	}

	// read flags
	this.Dup = ((flags >> 3) & 0x1) == 1
	this.Retain = (flags & 0x1) == 1
	this.QoS = (flags >> 1) & 0x3

	// check qos
	if !validQoS(this.QoS) {
		return total, fmt.Errorf("PUBLISH/Decode: Invalid QoS (%d).", this.QoS)
	}

	n := 0

	// read topic
	this.Topic, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	if this.QoS != 0 {
		// check buffer length
		if len(src) < total+2 {
			return total, fmt.Errorf("PUBLISH/Decode: Insufficient buffer size. Expecting %d, got %d.", total+2, len(src))
		}

		// read packet id
		this.PacketId = binary.BigEndian.Uint16(src[total:])
		total += 2
	}

	// calculate payload length
	l := int(rl) - (total - hl)

	if l > 0 {
		// check buffer length
		if len(src) < total+l {
			return total, fmt.Errorf("PUBLISH/Decode: Insufficient buffer size. Expecting %d, got %d.", total+l, len(src))
		}

		// read payload
		this.Payload = src[total : total+l]
		total += len(this.Payload)
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *PublishMessage) Encode(dst []byte) (int, error) {
	total := 0

	// check buffer length
	l := this.Len()
	if len(dst) < l {
		return total, fmt.Errorf("PUBLISH/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	// check topic length
	if len(this.Topic) == 0 {
		return total, fmt.Errorf("PUBLISH/Encode: Topic name is empty.")
	}

	flags := byte(0)

	// set dup flag
	if this.Dup {
		flags |= 0x8 // 00001000
	} else {
		flags &= 247 // 11110111
	}

	// set retain flag
	if this.Retain {
		flags |= 0x1 // 00000001
	} else {
		flags &= 254 // 11111110
	}

	// check qos
	if !validQoS(this.QoS) {
		return 0, fmt.Errorf("PUBLISH/Encode: Invalid QoS %d.", this.QoS)
	}

	// set qos
	flags = (flags & 249) | (this.QoS << 1) // 249 = 11111001

	// encode header
	n, err := this.header.encode(dst[total:], flags, this.msglen())
	total += n
	if err != nil {
		return total, err
	}

	// write topic
	n, err = writeLPBytes(dst[total:], this.Topic)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	if this.QoS != 0 {
		binary.BigEndian.PutUint16(dst[total:], this.PacketId)
		total += 2
	}

	// write payload
	copy(dst[total:], this.Payload)
	total += len(this.Payload)

	return total, nil
}

func (this *PublishMessage) msglen() int {
	total := 2 + len(this.Topic) + len(this.Payload)
	if this.QoS != 0 {
		total += 2
	}

	return total
}
