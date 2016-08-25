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

const maxRemainingLength int = 268435455 // bytes, or 256 MB

// Returns the length of the fixed header in bytes.
func headerLen(rl int) int {
	// packet type and flag byte
	total := 1

	if rl <= 127 {
		total++
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
func headerEncode(dst []byte, flags byte, rl int, tl int, t Type) (int, error) {
	total := 0

	// check buffer length
	if len(dst) < tl {
		return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", t, tl, len(dst))
	}

	// check remaining length
	if rl > maxRemainingLength || rl < 0 {
		return total, fmt.Errorf("[%s] remaining length (%d) out of bound (max %d, min 0)", t, rl, maxRemainingLength)
	}

	// check header length
	hl := headerLen(rl)
	if len(dst) < hl {
		return total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", t, hl, len(dst))
	}

	// write type and flags
	typeAndFlags := byte(t)<<4 | (t.defaultFlags() & 0xf)
	typeAndFlags |= flags
	dst[total] = typeAndFlags
	total++

	// write remaining length
	n := binary.PutUvarint(dst[total:], uint64(rl))
	total += n

	return total, nil
}

// Decodes the fixed header.
func headerDecode(src []byte, t Type) (int, byte, int, error) {
	total := 0

	// check buffer size
	if len(src) < 2 {
		return total, 0, 0, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", t, 2, len(src))
	}

	// read type and flags
	typeAndFlags := src[total : total+1]
	decodedType := Type(typeAndFlags[0] >> 4)
	flags := typeAndFlags[0] & 0x0f
	total++

	// check against static type
	if decodedType != t {
		return total, 0, 0, fmt.Errorf("[%s] invalid type %d", t, decodedType)
	}

	// check flags except for publish packets
	if t != PUBLISH && flags != t.defaultFlags() {
		return total, 0, 0, fmt.Errorf("[%s] invalid flags, expected %d, got %d", t, t.defaultFlags(), flags)
	}

	// read remaining length
	_rl, m := binary.Uvarint(src[total:])
	rl := int(_rl)
	total += m

	// check resulting remaining length
	if m <= 0 {
		return total, 0, 0, fmt.Errorf("[%s] error reading remaining length", t)
	}

	// check remaining buffer
	if rl > len(src[total:]) {
		return total, 0, 0, fmt.Errorf("[%s] remaining length (%d) is greater than remaining buffer (%d)", t, rl, len(src[total:]))
	}

	return total, flags, rl, nil
}
