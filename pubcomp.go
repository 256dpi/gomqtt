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

// The PUBCOMP Packet is the response to a PUBREL Packet. It is the fourth and
// final packet of the QoS 2 protocol exchange.
type PubcompMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubcompMessage)(nil)

// NewPubcompMessage creates a new PUBCOMP message.
func NewPubcompMessage() *PubcompMessage {
	msg := &PubcompMessage{}
	return msg
}

// Type return the messages message type.
func (this PubcompMessage) Type() MessageType {
	return PUBCOMP
}

// Len returns the byte length of the message.
func (this *PubcompMessage) Len() int {
	return identifiedMessageLen()
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *PubcompMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBCOMP)
	this.PacketId = pid
	return n, err
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *PubcompMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, this.PacketId, PUBCOMP)
}

// String returns a string representation of the message.
func (this PubcompMessage) String() string {
	return fmt.Sprintf("PUBCOMP: PacketId=%d", this.PacketId)
}
