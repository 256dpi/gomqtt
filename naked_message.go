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

type nakedMessage struct {
	header
}

// String returns a string representation of the message.
func (this nakedMessage) String() string {
	return fmt.Sprintf("Type=%q", this.Name())
}

// Name returns a string representation of the message type.
// TODO: Needed for proper documentation generation.
func (this *nakedMessage) Name() string {
	return this.header.Name()
}

// Len returns the byte length of the message.
func (this *nakedMessage) Len() int {
	return this.header.len(0)
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *nakedMessage) Decode(src []byte) (int, error) {
	// decode header
	hl, _, rl, err := this.header.decode(src)

	// check remaining length
	if rl != 0 {
		return hl, fmt.Errorf(this.Name() + "/Decode: Expected zero remaining length.")
	}

	return hl, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *nakedMessage) Encode(dst []byte) (int, error) {
	// encode header
	return this.header.encode(dst, 0, 0)
}
