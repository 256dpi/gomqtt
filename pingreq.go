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

// The PINGREQ Packet is sent from a Client to the Server. It can be used to:
// 1. Indicate to the Server that the Client is alive in the absence of any other
//    Control Packets being sent from the Client to the Server.
// 2. Request that the Server responds to confirm that it is alive.
// 3. Exercise the network to indicate that the Network Connection is active.
type PingreqMessage struct{}

var _ Message = (*PingreqMessage)(nil)

// NewPingreqMessage creates a new PINGREQ message.
func NewPingreqMessage() *PingreqMessage {
	return &PingreqMessage{}
}

// Type return the messages message type.
func (this PingreqMessage) Type() MessageType {
	return PINGREQ
}

// Len returns the byte length of the message.
func (this *PingreqMessage) Len() int {
	return nakedMessageLen()
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *PingreqMessage) Decode(src []byte) (int, error) {
	return nakedMessageDecode(src, PINGREQ)
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *PingreqMessage) Encode(dst []byte) (int, error) {
	return nakedMessageEncode(dst, PINGREQ)
}

// String returns a string representation of the message.
func (this PingreqMessage) String() string {
	return PINGREQ.String()
}
