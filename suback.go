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

// A SUBACK Packet is sent by the Server to the Client to confirm receipt and processing
// of a SUBSCRIBE Packet.
//
// A SUBACK Packet contains a list of return codes, that specify the maximum QoS level
// that was granted in each Subscription that was requested by the SUBSCRIBE.
type SubackMessage struct {
	header

	// The granted QoS levels for the requested subscriptions.
	ReturnCodes []byte

	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*SubackMessage)(nil)

// NewSubackMessage creates a new SUBACK message.
func NewSubackMessage() *SubackMessage {
	msg := &SubackMessage{}
	msg.Type = SUBACK
	return msg
}

// String returns a string representation of the message.
func (this SubackMessage) String() string {
	return fmt.Sprintf("%s: PacketId=%d ReturnCodes=%v", this.Type, this.PacketId, this.ReturnCodes)
}

// Len returns the byte length of the message.
func (this *SubackMessage) Len() int {
	ml := this.msglen()
	return this.header.len(ml) + ml
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *SubackMessage) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := this.header.decode(src[total:])
	total += hl
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("%s/Decode: Insufficient buffer size. Expecting %d, got %d.", this.Type, total+2, len(src))
	}

	// check remaining length
	if rl <= 2 {
		return total, fmt.Errorf("%s/Decode: Expected remaining length to be greater that 2, got.", this.Type, rl)
	}

	// read packet id
	this.PacketId = binary.BigEndian.Uint16(src[total:])
	total += 2

	// calculate number of return codes
	rcl := int(rl) - 2

	// check buffer length
	if len(src) < total+rcl {
		return total, fmt.Errorf("%s/Decode: Insufficient buffer size. Expecting %d, got %d.", this.Type, total+rcl, len(src))
	}

	// read return codes
	this.ReturnCodes = src[total : total+rcl]
	total += len(this.ReturnCodes)

	// validate return codes
	for i, code := range this.ReturnCodes {
		if !validQoS(code) && code != 0x80 {
			return total, fmt.Errorf("%s/Decode: Invalid return code %d for topic %d.", this.Type, code, i)
		}
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *SubackMessage) Encode(dst []byte) (int, error) {
	total := 0

	// check buffer length
	l := this.Len()
	if len(dst) < l {
		return total, fmt.Errorf("%s/Encode: Insufficient buffer size. Expecting %d, got %d.", this.Type, l, len(dst))
	}

	// check return codes
	for i, code := range this.ReturnCodes {
		if !validQoS(code) && code != 0x80 {
			return total, fmt.Errorf("%s/Encode: Invalid return code %d for topic %d.", this.Type, code, i)
		}
	}

	// encode header
	n, err := this.header.encode(dst[total:], 0, this.msglen())
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], this.PacketId)
	total += 2

	// write return codes
	copy(dst[total:], this.ReturnCodes)
	total += len(this.ReturnCodes)

	return total, nil
}

func (this *SubackMessage) msglen() int {
	return 2 + len(this.ReturnCodes)
}
