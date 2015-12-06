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

// An UnsubscribePacket is sent by the client to the server.
type UnsubscribePacket struct {
	// The topics to unsubscribe from.
	Topics [][]byte

	// The packet identifier.
	PacketID uint16
}

var _ Packet = (*UnsubscribePacket)(nil)

// NewUnsubscribePacket creates a new UnsubscribePacket.
func NewUnsubscribePacket() *UnsubscribePacket {
	return &UnsubscribePacket{}
}

// Type returns the packets type.
func (up UnsubscribePacket) Type() Type {
	return UNSUBSCRIBE
}

// String returns a string representation of the packet.
func (up UnsubscribePacket) String() string {
	s := "UNSUBSCRIBE:"

	for i, t := range up.Topics {
		s = fmt.Sprintf("%s Topic[%d]=%s", s, i, string(t))
	}

	return s
}

// Len returns the byte length of the encoded packet.
func (up *UnsubscribePacket) Len() int {
	ml := up.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
// The byte slice must not be modified during the duration of this packet being
// available since the byte slice never gets copied.
func (up *UnsubscribePacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src[total:], UNSUBSCRIBE)
	total += hl
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
	}

	// read packet id
	up.PacketID = binary.BigEndian.Uint16(src[total:])
	total += 2

	// prepare counter
	tl := int(rl) - 2

	// reset topics
	up.Topics = up.Topics[:0]

	for tl > 0 {
		// read topic
		t, n, err := readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		// append to list
		up.Topics = append(up.Topics, t)

		// decrement counter
		tl = tl - n - 1
	}

	// check for empty list
	if len(up.Topics) == 0 {
		return total, fmt.Errorf("Empty topic list")
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (up *UnsubscribePacket) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, up.len(), up.Len(), UNSUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], up.PacketID)
	total += 2

	for _, t := range up.Topics {
		// write topic
		n, err := writeLPBytes(dst[total:], t)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Returns the payload length.
func (up *UnsubscribePacket) len() int {
	// packet ID
	total := 2

	for _, t := range up.Topics {
		total += 2 + len(t)
	}

	return total
}
