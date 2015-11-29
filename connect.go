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
	version311Name = []byte{'M', 'Q', 'T', 'T'}
	version311Byte byte = 4
)

// After a Network Connection is established by a Client to a Server, the first Packet
// sent from the Client to the Server MUST be a CONNECT Packet [MQTT-3.1.0-1].
//
// A Client can only send the CONNECT Packet once over a Network Connection. The Server
// MUST process a second CONNECT Packet sent from a Client as a protocol violation and
// disconnect the Client [MQTT-3.1.0-2].  See section 4.8 for information about
// handling errors.
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
func (this ConnectMessage) Type() MessageType {
	return CONNECT
}

// String returns a string representation of the message.
func (this ConnectMessage) String() string {
	return fmt.Sprintf("CONNECT: KeepAlive=%d ClientId=%q WillTopic=%q WillPayload=%q Username=%q Password=%q",
		this.KeepAlive,
		this.ClientId,
		this.WillTopic,
		this.WillPayload,
		this.Username,
		this.Password,
	)
}

// Len returns the byte length of the message.
func (this *ConnectMessage) Len() int {
	ml := this.len()
	return headerLen(ml) + ml
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *ConnectMessage) Decode(src []byte) (int, error) {
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
		return total, fmt.Errorf("CONNECT/Decode: Insufficient buffer size. Expecting %d, got %d.", total+1, len(src))
	}

	// read version
	versionByte := src[total]
	total++

	// check protocol string and version
	if versionByte != version311Byte {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: Invalid protocol version (%d).", version311Byte)
	}

	// check protocol version string
	if (!bytes.Equal(protoName, version311Name)) {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: Invalid protocol version description (%s).", protoName)
	}

	// check buffer length
	if len(src) < total+1 {
		return total, fmt.Errorf("CONNECT/Decode: Insufficient buffer size. Expecting %d, got %d.", total+1, len(src))
	}

	// read connect flags
	connectFlags := src[total]
	total++

	// read existence flags
	usernameFlag := ((connectFlags >> 7) & 0x1) == 1
	passwordFlag := ((connectFlags >> 6) & 0x1) == 1
	willFlag := ((connectFlags >> 2) & 0x1) == 1

	// read other flags
	this.WillRetain = ((connectFlags >> 5) & 0x1) == 1
	this.WillQoS = (connectFlags >> 3) & 0x3
	this.CleanSession = ((connectFlags >> 1) & 0x1) == 1

	// check reserved bit
	if connectFlags&0x1 != 0 {
		return total, fmt.Errorf("CONNECT/Decode: Reserved bit 0 is not 0.")
	}

	// check will qos
	if !validQoS(this.WillQoS) {
		return total, fmt.Errorf("CONNECT/Decode: Invalid QoS level (%d) for will message.", this.WillQoS)
	}

	// check will flags
	if !willFlag && (this.WillRetain || this.WillQoS != 0) {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: If the Will Flag (%t) is set to 0 the Will QoS (%d) and Will Retain (%t) fields MUST be set to zero.", willFlag, this.WillQoS, this.WillRetain)
	}

	// check auth flags
	if !usernameFlag && passwordFlag {
		return total, fmt.Errorf("CONNECT/Decode: Password flag is set but Username flag is not set.")
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("CONNECT/Decode: Insufficient buffer size. Expecting %d, got %d.", total+2, len(src))
	}

	// read keep alive
	this.KeepAlive = binary.BigEndian.Uint16(src[total:])
	total += 2

	// read client id
	this.ClientId, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// if the client supplies a zero-byte ClientId, the Client MUST also set CleanSession to 1
	if len(this.ClientId) == 0 && !this.CleanSession {
		return total, fmt.Errorf("CONNECT/Decode: Protocol violation: Clean session must be 1 if client id is zero length.")
	}

	// read will topic and payload
	if willFlag {
		this.WillTopic, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		this.WillPayload, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// read username
	if usernameFlag {
		this.Username, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// read password
	if passwordFlag {
		this.Password, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *ConnectMessage) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, this.len(), this.Len(), CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// write version string
	n, err = writeLPBytes(dst[total:], version311Name)
	total += n
	if err != nil {
		return total, err
	}

	// write version value
	dst[total] = version311Byte
	total += 1

	var connectFlags byte

	// set username flag
	if len(this.Username) > 0 {
		connectFlags |= 128 // 10000000
	} else {
		connectFlags &= 127 // 01111111
	}

	// set password flag
	if len(this.Password) > 0 {
		connectFlags |= 64 // 01000000
	} else {
		connectFlags &= 191 // 10111111
	}

	// set will flag
	if len(this.WillTopic) > 0 {
		connectFlags |= 0x4 // 00000100

		if !validQoS(this.WillQoS) {
			return total, fmt.Errorf("CONNECT/Encode: Invalid Will QoS level %d.", this.WillQoS)
		}

		// set will qos flag
		connectFlags = (connectFlags & 231) | (this.WillQoS << 3) // 231 = 11100111

		// set will retain flag
		if this.WillRetain {
			connectFlags |= 32 // 00100000
		} else {
			connectFlags &= 223 // 11011111
		}

	} else {
		connectFlags &= 251 // 11111011
	}

	// set clean session flag
	if this.CleanSession {
		connectFlags |= 0x2 // 00000010
	} else {
		connectFlags &= 253 // 11111101
	}

	// write connect flags
	dst[total] = connectFlags
	total += 1

	// write keep alive
	binary.BigEndian.PutUint16(dst[total:], this.KeepAlive)
	total += 2

	// write client id
	n, err = writeLPBytes(dst[total:], this.ClientId)
	total += n
	if err != nil {
		return total, err
	}

	// write will topic and payload
	if len(this.WillTopic) > 0 {
		n, err = writeLPBytes(dst[total:], this.WillTopic)
		total += n
		if err != nil {
			return total, err
		}

		n, err = writeLPBytes(dst[total:], this.WillPayload)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write username
	if len(this.Username) > 0 {
		n, err = writeLPBytes(dst[total:], this.Username)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write password
	if len(this.Password) > 0 {
		n, err = writeLPBytes(dst[total:], this.Password)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// Returns the payload length.
func (this *ConnectMessage) len() int {
	total := 0

	// 2 bytes protocol name length
	// 4 bytes protocol name
	// 1 byte protocol version
	// 1 byte connect flags
	// 2 bytes keep alive timer
	total += 2 + 4 + 1 + 1 + 2

	// add the clientID length
	total += 2 + len(this.ClientId)

	// add the will topic and will message length
	if len(this.WillTopic) > 0 {
		total += 2 + len(this.WillTopic) + 2 + len(this.WillPayload)
	}

	// add the username length
	if len(this.Username) > 0 {
		total += 2 + len(this.Username)
	}

	// add the password length
	if len(this.Password) > 0 {
		total += 2 + len(this.Password)
	}

	return total
}
