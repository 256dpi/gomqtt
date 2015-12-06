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

// Len returns the byte length of the encoded packet.
func nakedPacketLen() int {
	return headerLen(0)
}

// Decodes a naked packet.
func nakedPacketDecode(src []byte, t Type) (int, error) {
	// decode header
	hl, _, rl, err := headerDecode(src, t)

	// check remaining length
	if rl != 0 {
		return hl, fmt.Errorf("%s/nakedPacketDecode: Expected zero remaining length", t)
	}

	return hl, err
}

// Encodes a naked packet.
func nakedPacketEncode(dst []byte, t Type) (int, error) {
	// encode header
	return headerEncode(dst, 0, 0, nakedPacketLen(), t)
}

// The DisconnectPacket is sent from the Client to the Server.
// It indicates that the Client is disconnecting cleanly.
type DisconnectPacket struct{}

var _ Packet = (*DisconnectPacket)(nil)

// NewDisconnectPacket creates a new DISCONNECT packet.
func NewDisconnectPacket() *DisconnectPacket {
	return &DisconnectPacket{}
}

// Type returns the packets type.
func (dm DisconnectPacket) Type() Type {
	return DISCONNECT
}

// Len returns the byte length of the encoded packet.
func (dm *DisconnectPacket) Len() int {
	return nakedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (dm *DisconnectPacket) Decode(src []byte) (int, error) {
	return nakedPacketDecode(src, DISCONNECT)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (dm *DisconnectPacket) Encode(dst []byte) (int, error) {
	return nakedPacketEncode(dst, DISCONNECT)
}

// String returns a string representation of the packet.
func (dm DisconnectPacket) String() string {
	return DISCONNECT.String()
}

// The PingreqPacket is sent from a Client to the Server.
type PingreqPacket struct{}

var _ Packet = (*PingreqPacket)(nil)

// NewPingreqPacket creates a new PINGREQ packet.
func NewPingreqPacket() *PingreqPacket {
	return &PingreqPacket{}
}

// Type returns the packets type.
func (pm PingreqPacket) Type() Type {
	return PINGREQ
}

// Len returns the byte length of the encoded packet.
func (pm *PingreqPacket) Len() int {
	return nakedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PingreqPacket) Decode(src []byte) (int, error) {
	return nakedPacketDecode(src, PINGREQ)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PingreqPacket) Encode(dst []byte) (int, error) {
	return nakedPacketEncode(dst, PINGREQ)
}

// String returns a string representation of the packet.
func (pm PingreqPacket) String() string {
	return PINGREQ.String()
}

// A PingrespPacket is sent by the Server to the Client in response to a
// PingreqPacket. It indicates that the Server is alive.
type PingrespPacket struct{}

var _ Packet = (*PingrespPacket)(nil)

// NewPingrespPacket creates a new PINGRESP packet.
func NewPingrespPacket() *PingrespPacket {
	return &PingrespPacket{}
}

// Type returns the packets type.
func (pm PingrespPacket) Type() Type {
	return PINGRESP
}

// Len returns the byte length of the encoded packet.
func (pm *PingrespPacket) Len() int {
	return nakedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PingrespPacket) Decode(src []byte) (int, error) {
	return nakedPacketDecode(src, PINGRESP)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PingrespPacket) Encode(dst []byte) (int, error) {
	return nakedPacketEncode(dst, PINGRESP)
}

// String returns a string representation of the packet.
func (pm PingrespPacket) String() string {
	return PINGRESP.String()
}
