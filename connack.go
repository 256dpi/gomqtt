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

// The CONNACK Packet is the packet sent by the Server in response to a CONNECT Packet
// received from a Client. The first packet sent from the Server to the Client MUST
// be a CONNACK Packet [MQTT-3.2.0-1].
//
// If the Client does not receive a CONNACK Packet from the Server within a reasonable
// amount of time, the Client SHOULD close the Network Connection. A "reasonable" amount
// of time depends on the type of application and the communications infrastructure.
type ConnackMessage struct {
	header

	// The Session Present flag enables a Client to establish whether the Client and
	// Server have a consistent view about whether there is already stored Session state.
	SessionPresent bool

	// If a well formed CONNECT Packet is received by the Server, but the Server is unable
	// to process it for some reason, then the Server SHOULD attempt to send a CONNACK packet
	// containing the appropriate non-zero Connect return code.
	ReturnCode ConnackCode
}

var _ Message = (*ConnackMessage)(nil)

// NewConnackMessage creates a new CONNACK message.
func NewConnackMessage() *ConnackMessage {
	msg := &ConnackMessage{}
	msg.messageType = CONNACK
	return msg
}

func (this ConnackMessage) Type() MessageType {
	return CONNACK
}

// String returns a string representation of the message.
func (this ConnackMessage) String() string {
	return fmt.Sprintf("CONNACK: SessionPresent=%t ReturnCode=%q", this.SessionPresent, this.ReturnCode)
}

// Len returns the byte length of the message.
func (this *ConnackMessage) Len() int {
	return this.header.len(2) + 2
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *ConnackMessage) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := this.header.decode(src)
	total += hl
	if err != nil {
		return total, err
	}

	// check remaining length
	if rl != 2 {
		return total, fmt.Errorf("CONNACK/Decode: Expected remaining length to be 2.")
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("CONNACK/Decode: Insufficient buffer size. Expecting %d, got %d.", total+2, len(src))
	}

	// read connack flags
	connackFlags := src[total]
	this.SessionPresent = connackFlags&0x1 == 1
	total++

	// check flags
	if connackFlags&254 != 0 {
		return 0, fmt.Errorf("CONNACK/Decode: Bits 7-1 in acknowledge flags byte (1) are not 0.")
	}

	// read return code
	this.ReturnCode = ConnackCode(src[total])
	total++

	// check return code
	if !this.ReturnCode.Valid() {
		return 0, fmt.Errorf("CONNACK/Decode: Invalid return code (%d)", this.ReturnCode)
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *ConnackMessage) Encode(dst []byte) (int, error) {
	total := 0

	// check buffer length
	l := this.Len()
	if len(dst) < l {
		return total, fmt.Errorf("CONNACK/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	// encode header
	n, err := this.header.encode(dst[total:], 0, 2)
	total += n
	if err != nil {
		return total, err
	}

	// set session present flag
	if this.SessionPresent {
		dst[total] = 1 // 00000001
	} else {
		dst[total] = 0 // 00000000
	}
	total++

	// check return code
	if !this.ReturnCode.Valid() {
		return total, fmt.Errorf("CONNACK/Encode: Invalid return code (%d).", this.ReturnCode)
	}

	// set return code
	dst[total] = byte(this.ReturnCode)
	total++

	return total, nil
}
