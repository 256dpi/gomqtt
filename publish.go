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
	// The message to publish.
	Message Message

	// If the Dup flag is set to false, it indicates that this is the first
	// occasion that the client or server has attempted to send this
	// PublishPacket. If the dup flag is set to true, it indicates that this
	// might be re-delivery of an earlier attempt to send the Packet.
	Dup bool

	// The packet identifier.
	PacketID uint16
}

// NewPublishPacket creates a new PublishPacket.
func NewPublishPacket() *PublishPacket {
	return &PublishPacket{}
}

// Type returns the packets type.
func (pp *PublishPacket) Type() Type {
	return PUBLISH
}

// String returns a string representation of the packet.
func (pp *PublishPacket) String() string {
	return fmt.Sprintf("<PublishPacket PacketID=%d Message=%s Dup=%t>",
		pp.PacketID, pp.Message.String(), pp.Dup)
}

// Len returns the byte length of the encoded packet.
func (pp *PublishPacket) Len() int {
	ml := pp.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
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
	pp.Message.Retain = (flags & 0x1) == 1
	pp.Message.QOS = (flags >> 1) & 0x3

	// check qos
	if !validQOS(pp.Message.QOS) {
		return total, fmt.Errorf("[%s] invalid QOS level (%d)", pp.Type(), pp.Message.QOS)
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", pp.Type(), total+2, len(src))
	}

	n := 0

	// read topic
	pp.Message.Topic, n, err = readLPString(src[total:], pp.Type())
	total += n
	if err != nil {
		return total, err
	}

	if pp.Message.QOS != 0 {
		// check buffer length
		if len(src) < total+2 {
			return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", pp.Type(), total+2, len(src))
		}

		// read packet id
		pp.PacketID = binary.BigEndian.Uint16(src[total:])
		total += 2

		// check packet id
		if pp.PacketID == 0 {
			return total, fmt.Errorf("[%s] packet id must be grater than zero", pp.Type())
		}
	}

	// calculate payload length
	l := int(rl) - (total - hl)

	// read payload
	if l > 0 {
		pp.Message.Payload = make([]byte, l)
		copy(pp.Message.Payload, src[total:total+l])
		total += len(pp.Message.Payload)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PublishPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check topic length
	if len(pp.Message.Topic) == 0 {
		return total, fmt.Errorf("[%s] topic name is empty", pp.Type())
	}

	flags := byte(0)

	// set dup flag
	if pp.Dup {
		flags |= 0x8 // 00001000
	} else {
		flags &= 247 // 11110111
	}

	// set retain flag
	if pp.Message.Retain {
		flags |= 0x1 // 00000001
	} else {
		flags &= 254 // 11111110
	}

	// check qos
	if !validQOS(pp.Message.QOS) {
		return 0, fmt.Errorf("[%s] invalid QOS level %d", pp.Type(), pp.Message.QOS)
	}

	// check packet id
	if pp.Message.QOS > 0 && pp.PacketID == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", pp.Type())
	}

	// set qos
	flags = (flags & 249) | (pp.Message.QOS << 1) // 249 = 11111001

	// encode header
	n, err := headerEncode(dst[total:], flags, pp.len(), pp.Len(), PUBLISH)
	total += n
	if err != nil {
		return total, err
	}

	// write topic
	n, err = writeLPString(dst[total:], pp.Message.Topic, pp.Type())
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	if pp.Message.QOS != 0 {
		binary.BigEndian.PutUint16(dst[total:], pp.PacketID)
		total += 2
	}

	// write payload
	copy(dst[total:], pp.Message.Payload)
	total += len(pp.Message.Payload)

	return total, nil
}

// Returns the payload length.
func (pp *PublishPacket) len() int {
	total := 2 + len(pp.Message.Topic) + len(pp.Message.Payload)
	if pp.Message.QOS != 0 {
		total += 2
	}

	return total
}
