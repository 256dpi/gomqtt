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

const maxRemainingLength int = 268435455 // bytes, or 256 MB

// fixed header
type header struct {
	// The message type of the message.
	Type MessageType

	// TODO: move to messages that have a packet Id?
	// Some messages need a packet ID.
	PacketId uint16
}

// String returns a string representation of the message.
func (this header) String() string {
	return fmt.Sprintf("Type=%q", this.Name())
}

// Name returns a string representation of the message type.
func (this *header) Name() string {
	return this.Type.Name()
}

// Returns the length of the fixed header in bytes.
func (this *header) len(rl int) int {
	// message type and flag byte
	total := 1

	if rl <= 127 {
		total += 1
	} else if rl <= 16383 {
		total += 2
	} else if rl <= 2097151 {
		total += 3
	} else {
		total += 4
	}

	return total
}

// Encodes the fixed header.
func (this *header) encode(dst []byte, flags byte, rl int) (int, error) {
	total := 0

	// check remaining length
	if rl > maxRemainingLength || rl < 0 {
		return total, fmt.Errorf(this.Name()+"/encode: remaining length (%d) out of bound (max %d, min 0)", rl, maxRemainingLength)
	}

	// check header length
	hl := this.len(rl)
	if len(dst) < hl {
		return total, fmt.Errorf(this.Name()+"/encode: Insufficient buffer size. Expecting %d, got %d.", hl, len(dst))
	}

	// validate message type
	if !this.Type.Valid() {
		return total, fmt.Errorf(this.Name()+"/encode: Invalid message type %d", this.Type)
	}

	// write type and flags
	typeAndFlags := byte(this.Type)<<4 | (this.Type.defaultFlags() & 0xf)
	typeAndFlags |= flags
	dst[total] = typeAndFlags
	total += 1

	// write remaining length
	n := binary.PutUvarint(dst[total:], uint64(rl))
	total += n

	return total, nil
}

// Decodes the fixed header.
func (this *header) decode(src []byte) (int, byte, int, error) {
	total := 0

	// check buffer size
	if len(src) < 2 {
		return total, 0, 0, fmt.Errorf(this.Name()+"/encode: Insufficient buffer size. Expecting %d, got %d.", 2, len(src))
	}

	// cache old type
	oldType := this.Type

	// read type and flags
	typeAndFlags := src[total : total+1]
	this.Type = MessageType(typeAndFlags[0] >> 4)
	flags := typeAndFlags[0] & 0x0f
	total++

	// check new type
	if !this.Type.Valid() {
		return total, 0, 0, fmt.Errorf(this.Name()+"/decode: Invalid message type %d.", this.Type)
	}

	// check against old type
	if oldType != this.Type {
		return total, 0, 0, fmt.Errorf(this.Name()+"/decode: Invalid message type %d. Expecting %d.", this.Type, oldType)
	}

	//TODO: check this in message implementation
	if this.Type != PUBLISH && flags != this.Type.defaultFlags() {
		return total, 0, 0, fmt.Errorf(this.Name()+"/decode: Invalid message (%d) flags. Expecting %d, got %d", this.Type, this.Type.defaultFlags(), flags)
	}

	//TODO: check this in message implementation
	if this.Type == PUBLISH && !validQoS((flags>>1)&0x3) {
		return total, 0, 0, fmt.Errorf(this.Name()+"/decode: Invalid QoS (%d) for PUBLISH message.", (flags>>1)&0x3)
	}

	// read remaining length
	_rl, m := binary.Uvarint(src[total:])
	rl := int(_rl)
	total += m

	// check resulting remaining length
	if rl > maxRemainingLength || rl < 0 {
		return total, 0, 0, fmt.Errorf(this.Name()+"/decode: Remaining length (%d) out of bound (max %d, min 0)", rl, maxRemainingLength)
	}

	// check remaining buffer
	if rl > len(src[total:]) {
		return total, 0, 0, fmt.Errorf(this.Name()+"/decode: Remaining length (%d) is greater than remaining buffer (%d)", rl, len(src[total:]))
	}

	return total, flags, rl, nil
}
