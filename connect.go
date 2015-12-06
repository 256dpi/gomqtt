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

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	version311Name      = []byte{'M', 'Q', 'T', 'T'}
	version311Byte byte = 4
)

// After a Network Connection is established by a Client to a Server, the first Packet
// sent from the Client to the Server MUST be a CONNECT Packet.
type ConnectMessage struct {
	// The clients client id.
	ClientId []byte

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

	// The qos of the will message.
	WillQoS byte

	// The retain setting of the will message.
	WillRetain bool
}

var _ Message = (*ConnectMessage)(nil)

// NewConnectMessage creates a new CONNECT message.
func NewConnectMessage() *ConnectMessage {
	return &ConnectMessage{CleanSession: true}
}

// Type return the messages message type.
func (cm ConnectMessage) Type() MessageType {
	return CONNECT
}

// String returns a string representation of the message.
func (cm ConnectMessage) String() string {
	return fmt.Sprintf("CONNECT: KeepAlive=%d ClientId=%q WillTopic=%q WillPayload=%q Username=%q Password=%q",
		cm.KeepAlive,
		cm.ClientId,
		cm.WillTopic,
		cm.WillPayload,
		cm.Username,
		cm.Password,
	)
}

// Len returns the byte length of the message.
func (cm *ConnectMessage) Len() int {
	ml := cm.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (cm *ConnectMessage) Decode(src []byte) (int, error) {
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
		return total, fmt.Errorf("CONNECT/Decode: Insufficient buffer size. Expecting %d, got %d", total+1, len(src))
	}

	// read version
	versionByte := src[total]
	total++

	// check protocol string and version
	if versionByte != version311Byte {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: Invalid protocol version (%d)", version311Byte)
	}

	// check protocol version string
	if !bytes.Equal(protoName, version311Name) {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: Invalid protocol version description (%s)", protoName)
	}

	// check buffer length
	if len(src) < total+1 {
		return total, fmt.Errorf("CONNECT/Decode: Insufficient buffer size. Expecting %d, got %d", total+1, len(src))
	}

	// read connect flags
	connectFlags := src[total]
	total++

	// read existence flags
	usernameFlag := ((connectFlags >> 7) & 0x1) == 1
	passwordFlag := ((connectFlags >> 6) & 0x1) == 1
	willFlag := ((connectFlags >> 2) & 0x1) == 1

	// read other flags
	cm.WillRetain = ((connectFlags >> 5) & 0x1) == 1
	cm.WillQoS = (connectFlags >> 3) & 0x3
	cm.CleanSession = ((connectFlags >> 1) & 0x1) == 1

	// check reserved bit
	if connectFlags&0x1 != 0 {
		return total, fmt.Errorf("CONNECT/Decode: Reserved bit 0 is not 0")
	}

	// check will qos
	if !validQoS(cm.WillQoS) {
		return total, fmt.Errorf("CONNECT/Decode: Invalid QoS level (%d) for will message", cm.WillQoS)
	}

	// check will flags
	if !willFlag && (cm.WillRetain || cm.WillQoS != 0) {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: If the Will Flag (%t) is set to 0 the Will QoS (%d) and Will Retain (%t) fields MUST be set to zero", willFlag, cm.WillQoS, cm.WillRetain)
	}

	// check auth flags
	if !usernameFlag && passwordFlag {
		return total, fmt.Errorf("CONNECT/Decode: Password flag is set but Username flag is not set")
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("CONNECT/Decode: Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
	}

	// read keep alive
	cm.KeepAlive = binary.BigEndian.Uint16(src[total:])
	total += 2

	// read client id
	cm.ClientId, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// if the client supplies a zero-byte ClientId, the Client MUST also set CleanSession to 1
	if len(cm.ClientId) == 0 && !cm.CleanSession {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: Clean session must be 1 if client id is zero length")
	}

	// read will topic and payload
	if willFlag {
		cm.WillTopic, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		cm.WillPayload, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// read username
	if usernameFlag {
		cm.Username, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// read password
	if passwordFlag {
		cm.Password, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (cm *ConnectMessage) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, cm.len(), cm.Len(), CONNECT)
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
	if len(cm.Username) > 0 {
		connectFlags |= 128 // 10000000
	} else {
		connectFlags &= 127 // 01111111
	}

	// set password flag
	if len(cm.Password) > 0 {
		connectFlags |= 64 // 01000000
	} else {
		connectFlags &= 191 // 10111111
	}

	// set will flag
	if len(cm.WillTopic) > 0 {
		connectFlags |= 0x4 // 00000100

		if !validQoS(cm.WillQoS) {
			return total, fmt.Errorf("CONNECT/Encode: Invalid Will QoS level %d", cm.WillQoS)
		}

		// set will qos flag
		connectFlags = (connectFlags & 231) | (cm.WillQoS << 3) // 231 = 11100111

		// set will retain flag
		if cm.WillRetain {
			connectFlags |= 32 // 00100000
		} else {
			connectFlags &= 223 // 11011111
		}

	} else {
		connectFlags &= 251 // 11111011
	}

	// set clean session flag
	if cm.CleanSession {
		connectFlags |= 0x2 // 00000010
	} else {
		connectFlags &= 253 // 11111101
	}

	// write connect flags
	dst[total] = connectFlags
	total++

	// write keep alive
	binary.BigEndian.PutUint16(dst[total:], cm.KeepAlive)
	total += 2

	// write client id
	n, err = writeLPBytes(dst[total:], cm.ClientId)
	total += n
	if err != nil {
		return total, err
	}

	// write will topic and payload
	if len(cm.WillTopic) > 0 {
		n, err = writeLPBytes(dst[total:], cm.WillTopic)
		total += n
		if err != nil {
			return total, err
		}

		n, err = writeLPBytes(dst[total:], cm.WillPayload)
		total += n
		if err != nil {
			return total, err
		}
	}

	if len(cm.Username) == 0 && len(cm.Password) > 0 {
		return total, fmt.Errorf("CONNECT/Encode: Protocol violation: Password set without username")
	}

	// write username
	if len(cm.Username) > 0 {
		n, err = writeLPBytes(dst[total:], cm.Username)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write password
	if len(cm.Password) > 0 {
		n, err = writeLPBytes(dst[total:], cm.Password)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Returns the payload length.
func (cm *ConnectMessage) len() int {
	total := 0

	// 2 bytes protocol name length
	// 4 bytes protocol name
	// 1 byte protocol version
	// 1 byte connect flags
	// 2 bytes keep alive timer
	total += 2 + 4 + 1 + 1 + 2

	// add the clientID length
	total += 2 + len(cm.ClientId)

	// add the will topic and will message length
	if len(cm.WillTopic) > 0 {
		total += 2 + len(cm.WillTopic) + 2 + len(cm.WillPayload)
	}

	// add the username length
	if len(cm.Username) > 0 {
		total += 2 + len(cm.Username)
	}

	// add the password length
	if len(cm.Password) > 0 {
		total += 2 + len(cm.Password)
	}

	return total
}
