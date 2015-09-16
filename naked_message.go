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

import "fmt"

// Len returns the byte length of the message.
func nakedMessageLen() int {
	return headerLen(0)
}

// Decodes a naked message.
func nakedMessageDecode(src []byte, mt MessageType) (int, error) {
	// decode header
	hl, _, rl, err := headerDecode(src, mt)

	// check remaining length
	if rl != 0 {
		return hl, fmt.Errorf("%s/nakedMessageDecode: Expected zero remaining length.", mt)
	}

	return hl, err
}

// Encodes a naked message.
func nakedMessageEncode(dst []byte, mt MessageType) (int, error) {
	// encode header
	return headerEncode(dst, 0, 0, mt)
}
