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

import "fmt"

// Returns the byte length of a naked packet.
func nakedPacketLen() int {
	return headerLen(0)
}

// Decodes a naked packet.
func nakedPacketDecode(src []byte, t Type) (int, error) {
	// decode header
	hl, _, rl, err := headerDecode(src, t)

	// check remaining length
	if rl != 0 {
		return hl, fmt.Errorf("[%s] expected zero remaining length", t)
	}

	return hl, err
}

// Encodes a naked packet.
func nakedPacketEncode(dst []byte, t Type) (int, error) {
	// encode header
	return headerEncode(dst, 0, 0, nakedPacketLen(), t)
}

// A DisconnectPacket is sent from the client to the server.
// It indicates that the client is disconnecting cleanly.
type DisconnectPacket struct{}

// NewDisconnectPacket creates a new DisconnectPacket.
func NewDisconnectPacket() *DisconnectPacket {
	return &DisconnectPacket{}
}

// Type returns the packets type.
func (dp *DisconnectPacket) Type() Type {
	return DISCONNECT
}

// Len returns the byte length of the encoded packet.
func (dp *DisconnectPacket) Len() int {
	return nakedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (dp *DisconnectPacket) Decode(src []byte) (int, error) {
	return nakedPacketDecode(src, DISCONNECT)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (dp *DisconnectPacket) Encode(dst []byte) (int, error) {
	return nakedPacketEncode(dst, DISCONNECT)
}

// String returns a string representation of the packet.
func (dp *DisconnectPacket) String() string {
	return "<DisconnectPacket>"
}

// A PingreqPacket is sent from a client to the server.
type PingreqPacket struct{}

// NewPingreqPacket creates a new PingreqPacket.
func NewPingreqPacket() *PingreqPacket {
	return &PingreqPacket{}
}

// Type returns the packets type.
func (pp *PingreqPacket) Type() Type {
	return PINGREQ
}

// Len returns the byte length of the encoded packet.
func (pp *PingreqPacket) Len() int {
	return nakedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PingreqPacket) Decode(src []byte) (int, error) {
	return nakedPacketDecode(src, PINGREQ)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PingreqPacket) Encode(dst []byte) (int, error) {
	return nakedPacketEncode(dst, PINGREQ)
}

// String returns a string representation of the packet.
func (pp *PingreqPacket) String() string {
	return "<PingreqPacket>"
}

// A PingrespPacket is sent by the server to the client in response to a
// PingreqPacket. It indicates that the server is alive.
type PingrespPacket struct{}

var _ Packet = (*PingrespPacket)(nil)

// NewPingrespPacket creates a new PingrespPacket.
func NewPingrespPacket() *PingrespPacket {
	return &PingrespPacket{}
}

// Type returns the packets type.
func (pp *PingrespPacket) Type() Type {
	return PINGRESP
}

// Len returns the byte length of the encoded packet.
func (pp *PingrespPacket) Len() int {
	return nakedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PingrespPacket) Decode(src []byte) (int, error) {
	return nakedPacketDecode(src, PINGRESP)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PingrespPacket) Encode(dst []byte) (int, error) {
	return nakedPacketEncode(dst, PINGRESP)
}

// String returns a string representation of the packet.
func (pp *PingrespPacket) String() string {
	return "<PingrespPacket>"
}
