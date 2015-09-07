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
)

// DetectMessage tries to detect the next message in a buffer. It returns a
// length greater than zero if the message has been detected as well as its
// MessageType.
func DetectMessage(src []byte) (int, MessageType) {
	// check for minimum size
	if len(src) < 2 {
		return 0, 0
	}

	// get type
	mt := MessageType(src[0] >> 4)

	// get remaining length
	_rl, n := binary.Uvarint(src[1:])
	rl := int(_rl)

	if n <= 0 {
		return 0, 0
	}

	// check remaining length
	if rl > maxRemainingLength || rl < 0 {
		return 0, 0
	}

	return 1 + n + rl, mt
}
