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

// Len returns the byte length of the message.
func (this *identifiedMessage) Len() int {
	return this.header.len(2) + 2
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *identifiedMessage) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := this.header.decode(src)
	total += hl
	if err != nil {
		return total, err
	}

	// check remaining length
	if rl != 2 {
		return total, fmt.Errorf(this.Name() + "/Decode: Expected remaining length to be 2.")
	}

	// check buffer length
	if len(src) < total + 2 {
		return total, fmt.Errorf(this.Name() + "/Decode: Insufficient buffer size. Expecting %d, got %d.", total + 2, len(src))
	}

	// read packet id
	this.PacketId = binary.BigEndian.Uint16(src[total:])
	total += 2

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *identifiedMessage) Encode(dst []byte) (int, error) {
	total := 0

	// check buffer length
	l := this.Len()
	if len(dst) < l {
		return total, fmt.Errorf(this.Name()+"/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	// encode header
	n, err := this.header.encode(dst[total:], 0, 2)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], this.PacketId)
	total += 2

	return total, nil
}
