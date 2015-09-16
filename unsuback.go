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

// The UNSUBACK Packet is sent by the Server to the Client to confirm receipt of an
// UNSUBSCRIBE Packet.
type UnsubackMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*UnsubackMessage)(nil)

// NewUnsubackMessage creates a new UNSUBACK message.
func NewUnsubackMessage() *UnsubackMessage {
	return &UnsubackMessage{}
}

// Type return the messages message type.
func (this UnsubackMessage) Type() MessageType {
	return UNSUBACK
}

// Len returns the byte length of the message.
func (this *UnsubackMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *UnsubackMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, UNSUBACK)
	this.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *UnsubackMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, this.PacketId, UNSUBACK)
}

// String returns a string representation of the message.
func (this UnsubackMessage) String() string {
	return fmt.Sprintf("UNSUBACK: PacketId=%d", this.PacketId)
}
