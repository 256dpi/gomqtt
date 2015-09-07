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

	ReturnCodes []byte
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
	return fmt.Sprintf("%s, Packet ID=%d, Return Codes=%v", this.header, this.PacketId, this.ReturnCodes)
}

func (this *SubackMessage) Len() int {
	ml := this.msglen()
	return this.header.len(ml) + ml
}

// Decode message from the supplied buffer.
func (this *SubackMessage) Decode(src []byte) (int, error) {
	total := 0

	hl, _, rl, err := this.header.decode(src[total:])
	total += hl
	if err != nil {
		return total, err
	}

	this.PacketId = binary.BigEndian.Uint16(src[total:])
	total += 2

	l := int(rl) - (total - hl)
	this.ReturnCodes = src[total : total+l]
	total += len(this.ReturnCodes)

	for i, code := range this.ReturnCodes {
		if code != 0x00 && code != 0x01 && code != 0x02 && code != 0x80 {
			return total, fmt.Errorf(this.Name()+"/Decode: Invalid return code %d for topic %d", code, i)
		}
	}

	return total, nil
}

// Encode message to the supplied buffer.
func (this *SubackMessage) Encode(dst []byte) (int, error) {
	for i, code := range this.ReturnCodes {
		if code != 0x00 && code != 0x01 && code != 0x02 && code != 0x80 {
			return 0, fmt.Errorf(this.Name()+"/Encode: Invalid return code %d for topic %d", code, i)
		}
	}

	l := this.Len()

	if len(dst) < l {
		return 0, fmt.Errorf(this.Name()+"/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	total := 0

	n, err := this.header.encode(dst[total:], 0, this.msglen())
	total += n
	if err != nil {
		return total, err
	}

	binary.BigEndian.PutUint16(dst[total:], this.PacketId)
	total += 2

	copy(dst[total:], this.ReturnCodes)
	total += len(this.ReturnCodes)

	return total, nil
}

func (this *SubackMessage) msglen() int {
	return 2 + len(this.ReturnCodes)
}
