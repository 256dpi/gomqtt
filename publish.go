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
	"fmt"
	"encoding/binary"
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
}

var _ Message = (*PublishMessage)(nil)

// NewPublishMessage creates a new PUBLISH message.
func NewPublishMessage() *PublishMessage {
	msg := &PublishMessage{}
	msg.Type = PUBLISH
	return msg
}

func (this PublishMessage) String() string {
	return fmt.Sprintf("%s, Topic=%q, Packet ID=%d, QoS=%d, Retained=%t, Dup=%t, Payload=%v",
		this.header, this.Topic, this.PacketId, this.QoS, this.Retain, this.Dup, this.Payload)
}

// Len returns the byte length of the PublishMessage.
func (this *PublishMessage) Len() int {
	ml := this.msglen()
	return this.header.len(ml) + ml
}

// Decode message from the supplied buffer.
func (this *PublishMessage) Decode(src []byte) (int, error) {
	total := 0

	hl, flags, rl, err := this.header.decode(src[total:])
	total += hl
	if err != nil {
		return total, err
	}

	this.Dup = ((flags >> 3) & 0x1) == 1
	this.Retain = (flags & 0x1) == 1
	this.QoS = (flags >> 1) & 0x3

	n := 0

	this.Topic, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// The packet identifier field is only present in the PUBLISH packets where the
	// QoS level is 1 or 2
	if this.QoS != 0 {
		this.PacketId = binary.BigEndian.Uint16(src[total:])
		total += 2
	}

	l := int(rl) - (total - hl)
	this.Payload = src[total : total+l]
	total += len(this.Payload)

	return total, nil
}

// Encode message to the supplied buffer.
func (this *PublishMessage) Encode(dst []byte) (int, error) {
	if len(this.Topic) == 0 {
		return 0, fmt.Errorf(this.Name() + "/Encode: Topic name is empty.")
	}


	if len(this.Payload) == 0 {
		return 0, fmt.Errorf(this.Name() + "/Encode: Payload is empty.")
	}

	l := this.Len()

	if len(dst) < l {
		return 0, fmt.Errorf(this.Name() + "/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	total := 0

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
	if !ValidQoS(this.QoS) {
		return 0, fmt.Errorf(this.Name() + "/Encode: Invalid QoS %d.", this.QoS)
	}

	// set qos
	flags = (flags & 249) | (this.QoS << 1) // 249 = 11111001

	n, err := this.header.encode(dst[total:], flags, this.msglen())
	total += n
	if err != nil {
		return total, err
	}

	n, err = writeLPBytes(dst[total:], this.Topic)
	total += n
	if err != nil {
		return total, err
	}

	// The packet identifier field is only present in the PUBLISH packets where the QoS level is 1 or 2
	if this.QoS != 0 {
		binary.BigEndian.PutUint16(dst[total:], this.PacketId)
		total += 2
	}

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
