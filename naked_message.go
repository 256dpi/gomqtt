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

// Len returns byte length of message.
func (this *nakedMessage) Len() int {
	return this.header.len(0)
}

// Decode message from supplied buffer.
func (this *nakedMessage) Decode(src []byte) (int, error) {
	hl, _, rl, err := this.header.decode(src)

	if(rl != 0) {
		return hl, fmt.Errorf(this.Name() + "/Decode: Expected zero remaining length.")
	}

	return hl, err
}

// Encode message to supplied buffer.
func (this *nakedMessage) Encode(dst []byte) (int, error) {
	return this.header.encode(dst, 0, 0)
}
