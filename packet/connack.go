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

// The ConnackCode represents the return code in a ConnackPacket.
type ConnackCode uint8

// All available ConnackCodes.
const (
	ConnectionAccepted ConnackCode = iota
	ErrInvalidProtocolVersion
	ErrIdentifierRejected
	ErrServerUnavailable
	ErrBadUsernameOrPassword
	ErrNotAuthorized
)

// Valid checks if the ConnackCode is valid.
func (cc ConnackCode) Valid() bool {
	return cc <= 5
}

// Error returns the corresponding error string for the ConnackCode.
func (cc ConnackCode) Error() string {
	switch cc {
	case ConnectionAccepted:
		return "Connection accepted"
	case ErrInvalidProtocolVersion:
		return "Connection refused, unacceptable protocol version"
	case ErrIdentifierRejected:
		return "Connection refused, identifier rejected"
	case ErrServerUnavailable:
		return "Connection refused, Server unavailable"
	case ErrBadUsernameOrPassword:
		return "Connection refused, bad user name or password"
	case ErrNotAuthorized:
		return "Connection refused, not authorized"
	}

	return "Unknown error"
}

// A ConnackPacket is sent by the server in response to a ConnectPacket
// received from a client.
type ConnackPacket struct {
	// The SessionPresent flag enables a client to establish whether the
	// client and server have a consistent view about whether there is already
	// stored session state.
	SessionPresent bool

	// If a well formed ConnectPacket is received by the server, but the server
	// is unable to process it for some reason, then the server should attempt
	// to send a ConnackPacket containing a non-zero ReturnCode.
	ReturnCode ConnackCode
}

// NewConnackPacket creates a new ConnackPacket.
func NewConnackPacket() *ConnackPacket {
	return &ConnackPacket{}
}

// Type returns the packets type.
func (cp *ConnackPacket) Type() Type {
	return CONNACK
}

// String returns a string representation of the packet.
func (cp *ConnackPacket) String() string {
	return fmt.Sprintf("<ConnackPacket SessionPresent=%t ReturnCode=%d>",
		cp.SessionPresent, cp.ReturnCode)
}

// Len returns the byte length of the encoded packet.
func (cp *ConnackPacket) Len() int {
	return headerLen(2) + 2
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (cp *ConnackPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, CONNACK)
	total += hl
	if err != nil {
		return total, err
	}

	// check remaining length
	if rl != 2 {
		return total, fmt.Errorf("[%s] expected remaining length to be 2", cp.Type())
	}

	// read connack flags
	connackFlags := src[total]
	cp.SessionPresent = connackFlags&0x1 == 1
	total++

	// check flags
	if connackFlags&254 != 0 {
		return 0, fmt.Errorf("[%s] bits 7-1 in acknowledge flags are not 0", cp.Type())
	}

	// read return code
	cp.ReturnCode = ConnackCode(src[total])
	total++

	// check return code
	if !cp.ReturnCode.Valid() {
		return 0, fmt.Errorf("[%s] invalid return code (%d)", cp.Type(), cp.ReturnCode)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (cp *ConnackPacket) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, cp.Len(), CONNACK)
	total += n
	if err != nil {
		return total, err
	}

	// set session present flag
	if cp.SessionPresent {
		dst[total] = 1 // 00000001
	} else {
		dst[total] = 0 // 00000000
	}
	total++

	// check return code
	if !cp.ReturnCode.Valid() {
		return total, fmt.Errorf("[%s] invalid return code (%d)", cp.Type(), cp.ReturnCode)
	}

	// set return code
	dst[total] = byte(cp.ReturnCode)
	total++

	return total, nil
}
