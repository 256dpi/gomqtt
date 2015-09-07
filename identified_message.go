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

type identifiedMessage struct {
	header
}

// String returns a string representation of the message.
func (this identifiedMessage) String() string {
	return fmt.Sprintf("%s, Packet ID=%d", this.header, this.PacketId)
}

// Name returns a string representation of the message type.
// TODO: Needed for proper documentation generation.
func (this *identifiedMessage) Name() string {
	return this.header.Name()
}

// Len returns byte length of message.
func (this *identifiedMessage) Len() int {
	return this.header.len(2) + 2
}

// Decode message from supplied buffer.
func (this *identifiedMessage) Decode(src []byte) (int, error) {
	total := 0

	hl, _, rl, err := this.header.decode(src[total:])

	if rl != 2 {
		return hl, fmt.Errorf(this.Name() + "/Decode: Expected remaining length to be 2.")
	}

	total += hl
	if err != nil {
		return total, err
	}

	this.PacketId = binary.BigEndian.Uint16(src[total:])
	total += 2

	return total, nil
}

// Encode message to supplied buffer.
func (this *identifiedMessage) Encode(dst []byte) (int, error) {
	l := this.Len()

	if len(dst) < l {
		return 0, fmt.Errorf(this.Name()+"/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	total := 0

	n, err := this.header.encode(dst[total:], 0, 2)
	total += n
	if err != nil {
		return total, err
	}

	binary.BigEndian.PutUint16(dst[total:], this.PacketId)
	total += 2

	return total, nil
}
