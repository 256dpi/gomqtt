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

//type identifiedMessage struct {
//	// Shared message identifier.
//	PacketId uint16
//}

// String returns a string representation of the message.
//func (this identifiedMessage) String() string {
//	return fmt.Sprintf("%s: PacketId=%d", mt, this.PacketId)
//}

// Len returns the byte length of the message.
func identifiedMessageLen() int {
	return headerLen(2) + 2
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func identifiedMessageDecode(src []byte, mt MessageType) (int, uint16, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, mt)
	total += hl
	if err != nil {
		return total, 0, err
	}

	// check remaining length
	if rl != 2 {
		return total, 0, fmt.Errorf("%s/Decode: Expected remaining length to be 2.", mt)
	}

	// check buffer length
	if len(src) < total+2 {
		return total, 0, fmt.Errorf("%s/Decode: Insufficient buffer size. Expecting %d, got %d.", mt, total+2, len(src))
	}

	// read packet id
	packetId := binary.BigEndian.Uint16(src[total:])
	total += 2

	return total, packetId, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func identifiedMessageEncode(dst []byte, packetId uint16, mt MessageType) (int, error) {
	total := 0

	// check buffer length
	l := identifiedMessageLen()
	if len(dst) < l {
		return total, fmt.Errorf("%s/Encode: Insufficient buffer size. Expecting %d, got %d.", mt, l, len(dst))
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, mt)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], packetId)
	total += 2

	return total, nil
}
