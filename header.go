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
	if rl > maxRemainingLength || rl < 0 {
		return 0, fmt.Errorf(this.Name()+"/Encode: remaining length (%d) out of bound (max %d, min 0)", rl, maxRemainingLength)
	}

	hl := this.len(rl)

	if len(dst) < hl {
		return 0, fmt.Errorf(this.Name()+"/Encode: Insufficient buffer size. Expecting %d, got %d.", hl, len(dst))
	}

	total := 0

	if rl > maxRemainingLength || rl < 0 {
		return total, fmt.Errorf(this.Name()+"/Encode: Remaining length (%d) out of bound (max %d, min 0)", rl, maxRemainingLength)
	}

	if !this.Type.Valid() {
		return total, fmt.Errorf(this.Name()+"/Encode: Invalid message type %d", this.Type)
	}

	typeAndFlags := byte(this.Type)<<4 | (this.Type.defaultFlags() & 0xf)
	typeAndFlags |= flags

	dst[total] = typeAndFlags
	total += 1

	n := binary.PutUvarint(dst[total:], uint64(rl))
	total += n

	return total, nil
}

// Decodes the fixed header.
func (this *header) decode(src []byte) (int, byte, int, error) {
	total := 0

	// cache old type
	oldType := this.Type

	// read type and flags
	typeAndFlags := src[total : total+1]
	this.Type = MessageType(typeAndFlags[0] >> 4)
	flags := typeAndFlags[0] & 0x0f

	// check new type
	if !this.Type.Valid() {
		return total, 0, 0, fmt.Errorf(this.Name()+"/Decode: Invalid message type %d.", this.Type)
	}

	// check against old type
	if oldType != this.Type {
		return total, 0, 0, fmt.Errorf(this.Name()+"/Decode: Invalid message type %d. Expecting %d.", this.Type, oldType)
	}

	//TODO: check this in message implementation
	if this.Type != PUBLISH && flags != this.Type.defaultFlags() {
		return total, 0, 0, fmt.Errorf(this.Name()+"/Decode: Invalid message (%d) flags. Expecting %d, got %d", this.Type, this.Type.defaultFlags(), flags)
	}

	//TODO: check this in message implementation
	if this.Type == PUBLISH && !ValidQoS((flags>>1)&0x3) {
		return total, 0, 0, fmt.Errorf(this.Name()+"/Decode: Invalid QoS (%d) for PUBLISH message.", (flags>>1)&0x3)
	}

	total++

	_rl, m := binary.Uvarint(src[total:])
	rl := int(_rl)
	total += m

	if rl > maxRemainingLength || rl < 0 {
		return total, 0, 0, fmt.Errorf(this.Name()+"/Decode: Remaining length (%d) out of bound (max %d, min 0)", rl, maxRemainingLength)
	}

	if rl > len(src[total:]) {
		return total, 0, 0, fmt.Errorf(this.Name()+"/Decode: Remaining length (%d) is greater than remaining buffer (%d)", rl, len(src[total:]))
	}

	return total, flags, rl, nil
}
