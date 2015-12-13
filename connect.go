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

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	version311Name      = []byte{'M', 'Q', 'T', 'T'}
	version311Byte byte = 4
)

// A ConnectPacket is sent by a client to the server after a network
// connection has been established.
type ConnectPacket struct {
	// The clients client id.
	ClientID []byte

	// The keep alive value.
	KeepAlive uint16

	// The authentication username.
	Username []byte

	// The authentication password.
	Password []byte

	// The clean session flag.
	CleanSession bool

	// The topic of the will message.
	WillTopic []byte

	// The payload of the will message.
	WillPayload []byte

	// The QOS of the will message.
	WillQOS byte

	// The retain flag of the will message.
	WillRetain bool
}

var _ Packet = (*ConnectPacket)(nil)

// NewConnectPacket creates a new ConnectPacket.
func NewConnectPacket() *ConnectPacket {
	return &ConnectPacket{CleanSession: true}
}

// Type returns the packets type.
func (cp ConnectPacket) Type() Type {
	return CONNECT
}

// String returns a string representation of the packet.
func (cp ConnectPacket) String() string {
	return fmt.Sprintf("CONNECT: ClientID=%q KeepAlive=%d Username=%q "+
		"Password=%q CleanSession=%t WillTopic=%q WillPayload=%q WillQOS=%d "+
		"WillRetain=%t",
		cp.ClientID,
		cp.KeepAlive,
		cp.Username,
		cp.Password,
		cp.CleanSession,
		cp.WillTopic,
		cp.WillPayload,
		cp.WillQOS,
		cp.WillRetain,
	)
}

// Len returns the byte length of the encoded packet.
func (cp *ConnectPacket) Len() int {
	ml := cp.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
// The byte slice must not be modified during the duration of this packet being
// available since the byte slice never gets copied.
func (cp *ConnectPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, _, err := headerDecode(src[total:], CONNECT)
	total += hl
	if err != nil {
		return total, err
	}

	// read protocol string
	protoName, n, err := readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+1 {
		return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+1, len(src))
	}

	// read version
	versionByte := src[total]
	total++

	// check protocol string and version
	if versionByte != version311Byte {
		return total, fmt.Errorf("Protocol violation: Invalid protocol version (%d)", version311Byte)
	}

	// check protocol version string
	if !bytes.Equal(protoName, version311Name) {
		return total, fmt.Errorf("Protocol violation: Invalid protocol version description (%s)", protoName)
	}

	// check buffer length
	if len(src) < total+1 {
		return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+1, len(src))
	}

	// read connect flags
	connectFlags := src[total]
	total++

	// read existence flags
	usernameFlag := ((connectFlags >> 7) & 0x1) == 1
	passwordFlag := ((connectFlags >> 6) & 0x1) == 1
	willFlag := ((connectFlags >> 2) & 0x1) == 1

	// read other flags
	cp.WillRetain = ((connectFlags >> 5) & 0x1) == 1
	cp.WillQOS = (connectFlags >> 3) & 0x3
	cp.CleanSession = ((connectFlags >> 1) & 0x1) == 1

	// check reserved bit
	if connectFlags&0x1 != 0 {
		return total, fmt.Errorf("Reserved bit 0 is not 0")
	}

	// check will qos
	if !validQOS(cp.WillQOS) {
		return total, fmt.Errorf("Invalid QOS level (%d) for will message", cp.WillQOS)
	}

	// check will flags
	if !willFlag && (cp.WillRetain || cp.WillQOS != 0) {
		return total, fmt.Errorf("Protocol violation: If the Will Flag (%t) is set to 0 the Will QOS (%d) and Will Retain (%t) fields MUST be set to zero", willFlag, cp.WillQOS, cp.WillRetain)
	}

	// check auth flags
	if !usernameFlag && passwordFlag {
		return total, fmt.Errorf("Password flag is set but Username flag is not set")
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
	}

	// read keep alive
	cp.KeepAlive = binary.BigEndian.Uint16(src[total:])
	total += 2

	// read client id
	cp.ClientID, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// if the client supplies a zero-byte clientID, the client must also set CleanSession to 1
	if len(cp.ClientID) == 0 && !cp.CleanSession {
		return total, fmt.Errorf("Protocol violation: Clean session must be 1 if client id is zero length")
	}

	// read will topic and payload
	if willFlag {
		cp.WillTopic, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		cp.WillPayload, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// read username
	if usernameFlag {
		cp.Username, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// read password
	if passwordFlag {
		cp.Password, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (cp *ConnectPacket) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, cp.len(), cp.Len(), CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// write version string, length has been checked beforehand
	n, _ = writeLPBytes(dst[total:], version311Name)
	total += n

	// write version value
	dst[total] = version311Byte
	total++

	var connectFlags byte

	// set username flag
	if len(cp.Username) > 0 {
		connectFlags |= 128 // 10000000
	} else {
		connectFlags &= 127 // 01111111
	}

	// set password flag
	if len(cp.Password) > 0 {
		connectFlags |= 64 // 01000000
	} else {
		connectFlags &= 191 // 10111111
	}

	// set will flag
	if len(cp.WillTopic) > 0 {
		connectFlags |= 0x4 // 00000100

		if !validQOS(cp.WillQOS) {
			return total, fmt.Errorf("Invalid Will QOS level %d", cp.WillQOS)
		}

		// set will qos flag
		connectFlags = (connectFlags & 231) | (cp.WillQOS << 3) // 231 = 11100111

		// set will retain flag
		if cp.WillRetain {
			connectFlags |= 32 // 00100000
		} else {
			connectFlags &= 223 // 11011111
		}

	} else {
		connectFlags &= 251 // 11111011
	}

	// set clean session flag
	if cp.CleanSession {
		connectFlags |= 0x2 // 00000010
	} else {
		connectFlags &= 253 // 11111101
	}

	// write connect flags
	dst[total] = connectFlags
	total++

	// write keep alive
	binary.BigEndian.PutUint16(dst[total:], cp.KeepAlive)
	total += 2

	// write client id
	n, err = writeLPBytes(dst[total:], cp.ClientID)
	total += n
	if err != nil {
		return total, err
	}

	// write will topic and payload
	if len(cp.WillTopic) > 0 {
		n, err = writeLPBytes(dst[total:], cp.WillTopic)
		total += n
		if err != nil {
			return total, err
		}

		n, err = writeLPBytes(dst[total:], cp.WillPayload)
		total += n
		if err != nil {
			return total, err
		}
	}

	if len(cp.Username) == 0 && len(cp.Password) > 0 {
		return total, fmt.Errorf("Protocol violation: Password set without username")
	}

	// write username
	if len(cp.Username) > 0 {
		n, err = writeLPBytes(dst[total:], cp.Username)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write password
	if len(cp.Password) > 0 {
		n, err = writeLPBytes(dst[total:], cp.Password)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Returns the payload length.
func (cp *ConnectPacket) len() int {
	total := 0

	// 2 bytes protocol name length
	// 4 bytes protocol name
	// 1 byte protocol version
	// 1 byte connect flags
	// 2 bytes keep alive timer
	total += 2 + 4 + 1 + 1 + 2

	// add the clientID length
	total += 2 + len(cp.ClientID)

	// add the will topic and will message length
	if len(cp.WillTopic) > 0 {
		total += 2 + len(cp.WillTopic) + 2 + len(cp.WillPayload)
	}

	// add the username length
	if len(cp.Username) > 0 {
		total += 2 + len(cp.Username)
	}

	// add the password length
	if len(cp.Password) > 0 {
		total += 2 + len(cp.Password)
	}

	return total
}
