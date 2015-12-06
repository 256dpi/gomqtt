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

// A PublishPacket is sent from a client to a server or from server to a client
// to transport an application message.
type PublishPacket struct {
	// The Topic of the message.
	Topic []byte

	// The Payload of the message.
	Payload []byte

	// The QOS indicates the level of assurance for delivery.
	QOS byte

	// If the Retain flag is set to true, in a PublishPacket sent by a client
	// to a server, the server must store the application message and its QOS,
	// so that it can be delivered to future subscribers whose subscriptions
	// match its topic name.
	Retain bool

	// If the Dup flag is set to false, it indicates that this is the first
	// occasion that the client or server has attempted to send this
	// PublishPacket. If the dup flag is set to true, it indicates that this
	// might be re-delivery of an earlier attempt to send the Packet.
	Dup bool

	// The packet identifier.
	PacketID uint16
}

var _ Packet = (*PublishPacket)(nil)

// NewPublishPacket creates a new PublishPacket.
func NewPublishPacket() *PublishPacket {
	return &PublishPacket{}
}

// Type returns the packets type.
func (pp PublishPacket) Type() Type {
	return PUBLISH
}

// String returns a string representation of the packet.
func (pp PublishPacket) String() string {
	return fmt.Sprintf("PUBLISH: Topic=%q PacketID=%d QOS=%d Retained=%t Dup=%t Payload=%v",
		pp.Topic, pp.PacketID, pp.QOS, pp.Retain, pp.Dup, pp.Payload)
}

// Len returns the byte length of the encoded packet.
func (pp *PublishPacket) Len() int {
	ml := pp.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
// The byte slice must not be modified during the duration of this packet being
// available since the byte slice never gets copied.
func (pp *PublishPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, flags, rl, err := headerDecode(src[total:], PUBLISH)
	total += hl
	if err != nil {
		return total, err
	}

	// read flags
	pp.Dup = ((flags >> 3) & 0x1) == 1
	pp.Retain = (flags & 0x1) == 1
	pp.QOS = (flags >> 1) & 0x3

	// check qos
	if !validQOS(pp.QOS) {
		return total, fmt.Errorf("Invalid QOS level (%d)", pp.QOS)
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
	}

	n := 0

	// read topic
	pp.Topic, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	if pp.QOS != 0 {
		// check buffer length
		if len(src) < total+2 {
			return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
		}

		// read packet id
		pp.PacketID = binary.BigEndian.Uint16(src[total:])
		total += 2
	}

	// calculate payload length
	l := int(rl) - (total - hl)

	// read payload
	if l > 0 {
		pp.Payload = src[total : total+l]
		total += len(pp.Payload)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PublishPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check topic length
	if len(pp.Topic) == 0 {
		return total, fmt.Errorf("Topic name is empty")
	}

	flags := byte(0)

	// set dup flag
	if pp.Dup {
		flags |= 0x8 // 00001000
	} else {
		flags &= 247 // 11110111
	}

	// set retain flag
	if pp.Retain {
		flags |= 0x1 // 00000001
	} else {
		flags &= 254 // 11111110
	}

	// check qos
	if !validQOS(pp.QOS) {
		return 0, fmt.Errorf("Invalid QOS level %d", pp.QOS)
	}

	// set qos
	flags = (flags & 249) | (pp.QOS << 1) // 249 = 11111001

	// encode header
	n, err := headerEncode(dst[total:], flags, pp.len(), pp.Len(), PUBLISH)
	total += n
	if err != nil {
		return total, err
	}

	// write topic
	n, err = writeLPBytes(dst[total:], pp.Topic)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	if pp.QOS != 0 {
		binary.BigEndian.PutUint16(dst[total:], pp.PacketID)
		total += 2
	}

	// write payload
	copy(dst[total:], pp.Payload)
	total += len(pp.Payload)

	return total, nil
}

// Returns the payload length.
func (pp *PublishPacket) len() int {
	total := 2 + len(pp.Topic) + len(pp.Payload)
	if pp.QOS != 0 {
		total += 2
	}

	return total
}
