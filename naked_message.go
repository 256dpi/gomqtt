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
		return hl, fmt.Errorf("%s/nakedMessageDecode: Expected zero remaining length", mt)
	}

	return hl, err
}

// Encodes a naked message.
func nakedMessageEncode(dst []byte, mt MessageType) (int, error) {
	// encode header
	return headerEncode(dst, 0, 0, nakedMessageLen(), mt)
}

// The DISCONNECT Packet is the final Control Packet sent from the Client to the Server.
// It indicates that the Client is disconnecting cleanly.
type DisconnectMessage struct{}

var _ Message = (*DisconnectMessage)(nil)

// NewDisconnectMessage creates a new DISCONNECT message.
func NewDisconnectMessage() *DisconnectMessage {
	return &DisconnectMessage{}
}

// Type return the messages message type.
func (this DisconnectMessage) Type() MessageType {
	return DISCONNECT
}

// Len returns the byte length of the message.
func (this *DisconnectMessage) Len() int {
	return nakedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *DisconnectMessage) Decode(src []byte) (int, error) {
	return nakedMessageDecode(src, DISCONNECT)
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (this *DisconnectMessage) Encode(dst []byte) (int, error) {
	return nakedMessageEncode(dst, DISCONNECT)
}

// String returns a string representation of the message.
func (this DisconnectMessage) String() string {
	return DISCONNECT.String()
}

// The PINGREQ Packet is sent from a Client to the Server.
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

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *PingreqMessage) Decode(src []byte) (int, error) {
	return nakedMessageDecode(src, PINGREQ)
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (this *PingreqMessage) Encode(dst []byte) (int, error) {
	return nakedMessageEncode(dst, PINGREQ)
}

// String returns a string representation of the message.
func (this PingreqMessage) String() string {
	return PINGREQ.String()
}

// A PINGRESP Packet is sent by the Server to the Client in response to a PINGREQ
// Packet. It indicates that the Server is alive.
type PingrespMessage struct{}

var _ Message = (*PingrespMessage)(nil)

// NewPingrespMessage creates a new PINGRESP message.
func NewPingrespMessage() *PingrespMessage {
	return &PingrespMessage{}
}

// Type return the messages message type.
func (this PingrespMessage) Type() MessageType {
	return PINGRESP
}

// Len returns the byte length of the message.
func (this *PingrespMessage) Len() int {
	return nakedMessageLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *PingrespMessage) Decode(src []byte) (int, error) {
	return nakedMessageDecode(src, PINGRESP)
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (this *PingrespMessage) Encode(dst []byte) (int, error) {
	return nakedMessageEncode(dst, PINGRESP)
}

// String returns a string representation of the message.
func (this PingrespMessage) String() string {
	return PINGRESP.String()
}
