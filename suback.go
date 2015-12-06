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

// A SubackPacket is sent by the server to the client to confirm receipt and
// processing of a SubscribePacket. The SubackPacket contains a list of return
// codes, that specify the maximum QOS levels that have been granted.
type SubackPacket struct {
	// The granted QOS levels for the requested subscriptions.
	ReturnCodes []byte

	// The packet identifier.
	PacketID uint16
}

var _ Packet = (*SubackPacket)(nil)

// NewSubackPacket creates a new SubackPacket.
func NewSubackPacket() *SubackPacket {
	return &SubackPacket{}
}

// Type returns the packets type.
func (sp SubackPacket) Type() Type {
	return SUBACK
}

// String returns a string representation of the packet.
func (sp SubackPacket) String() string {
	return fmt.Sprintf("SUBACK: PacketID=%d ReturnCodes=%v", sp.PacketID, sp.ReturnCodes)
}

// Len returns the byte length of the encoded packet.
func (sp *SubackPacket) Len() int {
	ml := sp.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
// The byte slice must not be modified during the duration of this packet being
// available since the byte slice never gets copied.
func (sp *SubackPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src[total:], SUBACK)
	total += hl
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
	}

	// check remaining length
	if rl <= 2 {
		return total, fmt.Errorf("Expected remaining length to be greater than 2, got %d", rl)
	}

	// read packet id
	sp.PacketID = binary.BigEndian.Uint16(src[total:])
	total += 2

	// calculate number of return codes
	rcl := int(rl) - 2

	// read return codes
	sp.ReturnCodes = src[total : total+rcl]
	total += len(sp.ReturnCodes)

	// validate return codes
	for i, code := range sp.ReturnCodes {
		if !validQOS(code) && code != QOSFailure {
			return total, fmt.Errorf("Invalid return code %d for topic %d", code, i)
		}
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (sp *SubackPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check return codes
	for i, code := range sp.ReturnCodes {
		if !validQOS(code) && code != QOSFailure {
			return total, fmt.Errorf("Invalid return code %d for topic %d", code, i)
		}
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, sp.len(), sp.Len(), SUBACK)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], sp.PacketID)
	total += 2

	// write return codes
	copy(dst[total:], sp.ReturnCodes)
	total += len(sp.ReturnCodes)

	return total, nil
}

// Returns the payload length.
func (sp *SubackPacket) len() int {
	return 2 + len(sp.ReturnCodes)
}
