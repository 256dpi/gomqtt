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

// An UNSUBSCRIBE Packet is sent by the Client to the Server, to unsubscribe from topics.
type UnsubscribeMessage struct {
	header

	Topics [][]byte
}

var _ Message = (*UnsubscribeMessage)(nil)

// NewUnsubscribeMessage creates a new UNSUBSCRIBE message.
func NewUnsubscribeMessage() *UnsubscribeMessage {
	msg := &UnsubscribeMessage{}
	msg.Type = UNSUBSCRIBE
	return msg
}

// String returns a string representation of the message.
func (this UnsubscribeMessage) String() string {
	msgstr := fmt.Sprintf("%s", this.header)

	for i, t := range this.Topics {
		msgstr = fmt.Sprintf("%s, Topic%d=%s", msgstr, i, string(t))
	}

	return msgstr
}

// Len returns the byte length of the message.
func (this *UnsubscribeMessage) Len() int {
	ml := this.msglen()
	return this.header.len(ml) + ml
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *UnsubscribeMessage) Decode(src []byte) (int, error) {
	total := 0

	hl, _, rl, err := this.header.decode(src[total:])
	total += hl
	if err != nil {
		return total, err
	}

	this.PacketId = binary.BigEndian.Uint16(src[total:])
	total += 2

	remlen := int(rl) - (total - hl)
	for remlen > 0 {
		t, n, err := readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		this.Topics = append(this.Topics, t)
		remlen = remlen - n - 1
	}

	if len(this.Topics) == 0 {
		return 0, fmt.Errorf(this.Name() + "/Decode: Empty topic list")
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *UnsubscribeMessage) Encode(dst []byte) (int, error) {
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

	for _, t := range this.Topics {
		n, err := writeLPBytes(dst[total:], t)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (this *UnsubscribeMessage) msglen() int {
	// packet ID
	total := 2

	for _, t := range this.Topics {
		total += 2 + len(t)
	}

	return total
}
