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
	// 0x3
	ProtocolV3Name = []byte{'M', 'Q', 'I', 's', 'd', 'p'}

	// 0x4
	ProtocolV4Name = []byte{'M', 'Q', 'T', 'T'}
)

// After a Network Connection is established by a Client to a Server, the first Packet
// sent from the Client to the Server MUST be a CONNECT Packet [MQTT-3.1.0-1].
//
// A Client can only send the CONNECT Packet once over a Network Connection. The Server
// MUST process a second CONNECT Packet sent from a Client as a protocol violation and
// disconnect the Client [MQTT-3.1.0-2].  See section 4.8 for information about
// handling errors.
type ConnectMessage struct {
	header

	Version   byte
	ClientId  []byte
	KeepAlive uint16

	Username     []byte
	Password     []byte
	CleanSession bool

	WillTopic   []byte
	WillMessage []byte
	WillQoS     byte
	WillRetain  bool

	//TODO: add documentationt to fields
}

var _ Message = (*ConnectMessage)(nil)

// NewConnectMessage creates a new CONNECT message.
func NewConnectMessage() *ConnectMessage {
	msg := &ConnectMessage{}
	msg.Type = CONNECT
	return msg
}

// String returns a string representation of the CONNECT message
func (this ConnectMessage) String() string {
	return fmt.Sprintf("%s, Version=%d, KeepAlive=%d, Client ID=%q, Will Topic=%q, Will Message=%q, Username=%q, Password=%q",
		this.header,
		this.Version,
		this.KeepAlive,
		this.ClientId,
		this.WillTopic,
		this.WillMessage,
		this.Username,
		this.Password,
	)
}

// Len returns the byte length of the CONNECT message.
func (this *ConnectMessage) Len() int {
	ml := this.msglen()
	return this.header.len(ml) + ml
}

// Decode decodes the CONNECT message from the supplied buffer.
func (this *ConnectMessage) Decode(src []byte) (int, error) {
	total := 0

	hl, _, _, err := this.header.decode(src[total:])
	if err != nil {
		return total + hl, err
	}
	total += hl

	if hl, err = this.decodeMessage(src[total:]); err != nil {
		return total + hl, err
	}
	total += hl

	return total, nil
}

// Encode encodes the CONNECT message in the supplied buffer.
func (this *ConnectMessage) Encode(dst []byte) (int, error) {
	if this.Version != 0x3 && this.Version != 0x4 {
		return 0, fmt.Errorf(this.Name()+"/Encode: Protocol violation: Invalid Protocol Version (%d) ", this.Version)
	}

	l := this.Len()

	if len(dst) < l {
		return 0, fmt.Errorf(this.Name()+"/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	total := 0

	n, err := this.header.encode(dst[total:], 0, this.msglen())
	total += n
	if err != nil {
		return total, err
	}

	n, err = this.encodeMessage(dst[total:])
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (this *ConnectMessage) encodeMessage(dst []byte) (int, error) {
	total := 0

	// write 0x3 name
	if this.Version == 0x3 {
		// write version string
		n, err := writeLPBytes(dst[total:], ProtocolV3Name)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write 0x4 name
	if this.Version == 0x4 {
		// write version string
		n, err := writeLPBytes(dst[total:], ProtocolV4Name)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write version value
	dst[total] = this.Version
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

		if !ValidQoS(this.WillQoS) {
			return total, fmt.Errorf(this.Name()+"/Encode: Invalid will QoS level %d", this.WillQoS)
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
	n, err := writeLPBytes(dst[total:], this.ClientId)
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

		n, err = writeLPBytes(dst[total:], this.WillMessage)
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

func (this *ConnectMessage) decodeMessage(src []byte) (int, error) {
	total := 0

	// read protocol string
	protoName, n, err := readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// read version
	this.Version = src[total]
	total++

	// check protocol string and version
	if this.Version != 0x3 && this.Version != 0x4 {
		return total, fmt.Errorf(this.Name()+"/decodeMessage: Protocol violation: Invalid Protocol version (%d) ", this.Version)
	}

	// check protocol version string
	if (this.Version == 0x3 && !bytes.Equal(protoName, ProtocolV3Name)) || (this.Version == 0x4 && !bytes.Equal(protoName, ProtocolV4Name)) {
		return total, fmt.Errorf(this.Name()+"/decodeMessage: Protocol violation: Invalid Protocol version description (%s) ", protoName)
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
		return total, fmt.Errorf(this.Name() + "/decodeMessage: Connect Flags reserved bit 0 is not 0")
	}

	// check will qos
	if !ValidQoS(this.WillQoS) {
		return total, fmt.Errorf(this.Name()+"/decodeMessage: Invalid QoS level (%d) for %s message", this.WillQoS, this.Name())
	}

	// check will flags
	if !willFlag && (this.WillRetain || this.WillQoS != 0) {
		return total, fmt.Errorf(this.Name()+"/decodeMessage: Protocol violation: If the Will Flag (%t) is set to 0 the Will QoS (%d) and Will Retain (%t) fields MUST be set to zero", willFlag, this.WillQoS, this.WillRetain)
	}

	// check auth flags
	if !usernameFlag && passwordFlag {
		return total, fmt.Errorf(this.Name() + "/decodeMessage: Password flag is set but Username flag is not set")
	}

	// check buffer length
	if len(src[total:]) < 2 {
		return 0, fmt.Errorf(this.Name()+"/decodeMessage: Insufficient buffer size. Expecting %d, got %d.", 2, len(src[total:]))
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
		return total, fmt.Errorf(this.Name() + "/decodeMessage: Protocol violation: Clean session must be 1 if client id is zero length.")
	}

	// read will topic and payload
	if willFlag {
		this.WillTopic, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		this.WillMessage, n, err = readLPBytes(src[total:])
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

func (this *ConnectMessage) msglen() int {
	total := 0

	vl := 0
	if this.Version == 0x3 {
		vl = 6
	} else if this.Version == 0x4 {
		vl = 4
	} else {
		return total
	}

	// 2 bytes protocol name length
	// n bytes protocol name
	// 1 byte protocol version
	// 1 byte connect flags
	// 2 bytes keep alive timer
	total += 2 + vl + 1 + 1 + 2

	// add the clientID length
	total += 2 + len(this.ClientId)

	// add the will topic and will message length
	if len(this.WillTopic) > 0 {
		total += 2 + len(this.WillTopic) + 2 + len(this.WillMessage)
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
