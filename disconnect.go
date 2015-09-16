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

// The DISCONNECT Packet is the final Control Packet sent from the Client to the Server.
// It indicates that the Client is disconnecting cleanly.
type DisconnectMessage struct{}

var _ Message = (*DisconnectMessage)(nil)

// NewDisconnectMessage creates a new DISCONNECT message.
func NewDisconnectMessage() *DisconnectMessage {
	msg := &DisconnectMessage{}
	return msg
}

// Type return the messages message type.
func (this DisconnectMessage) Type() MessageType {
	return DISCONNECT
}

// Len returns the byte length of the message.
func (this *DisconnectMessage) Len() int {
	return nakedMessageLen()
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *DisconnectMessage) Decode(src []byte) (int, error) {
	return nakedMessageDecode(src, DISCONNECT)
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *DisconnectMessage) Encode(dst []byte) (int, error) {
	return nakedMessageEncode(dst, DISCONNECT)
}

// String returns a string representation of the message.
func (this DisconnectMessage) String() string {
	return DISCONNECT.String()
}
