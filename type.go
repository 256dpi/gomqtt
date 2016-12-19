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

// Type represents the MQTT packet types.
type Type byte

// All packet types.
const (
	_ Type = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

// String returns the type as a string.
func (t Type) String() string {
	switch t {
	case CONNECT:
		return "Connect"
	case CONNACK:
		return "Connack"
	case PUBLISH:
		return "Publish"
	case PUBACK:
		return "Puback"
	case PUBREC:
		return "Pubrec"
	case PUBREL:
		return "Pubrel"
	case PUBCOMP:
		return "Pubcomp"
	case SUBSCRIBE:
		return "Subscribe"
	case SUBACK:
		return "Suback"
	case UNSUBSCRIBE:
		return "Unsubscribe"
	case UNSUBACK:
		return "Unsuback"
	case PINGREQ:
		return "Pingreq"
	case PINGRESP:
		return "Pingresp"
	case DISCONNECT:
		return "Disconnect"
	}

	return "Unknown"
}

// DefaultFlags returns the default flag values for the packet type, as defined
// by the MQTT spec, except for PUBLISH.
func (t Type) defaultFlags() byte {
	switch t {
	case CONNECT:
		return 0
	case CONNACK:
		return 0
	case PUBACK:
		return 0
	case PUBREC:
		return 0
	case PUBREL:
		return 2 // 00000010
	case PUBCOMP:
		return 0
	case SUBSCRIBE:
		return 2 // 00000010
	case SUBACK:
		return 0
	case UNSUBSCRIBE:
		return 2 // 00000010
	case UNSUBACK:
		return 0
	case PINGREQ:
		return 0
	case PINGRESP:
		return 0
	case DISCONNECT:
		return 0
	}

	return 0
}

// New creates a new packet based on the type. It is a shortcut to call one of
// the New*Packet functions. An error is returned if the type is invalid.
func (t Type) New() (Packet, error) {
	switch t {
	case CONNECT:
		return NewConnectPacket(), nil
	case CONNACK:
		return NewConnackPacket(), nil
	case PUBLISH:
		return NewPublishPacket(), nil
	case PUBACK:
		return NewPubackPacket(), nil
	case PUBREC:
		return NewPubrecPacket(), nil
	case PUBREL:
		return NewPubrelPacket(), nil
	case PUBCOMP:
		return NewPubcompPacket(), nil
	case SUBSCRIBE:
		return NewSubscribePacket(), nil
	case SUBACK:
		return NewSubackPacket(), nil
	case UNSUBSCRIBE:
		return NewUnsubscribePacket(), nil
	case UNSUBACK:
		return NewUnsubackPacket(), nil
	case PINGREQ:
		return NewPingreqPacket(), nil
	case PINGRESP:
		return NewPingrespPacket(), nil
	case DISCONNECT:
		return NewDisconnectPacket(), nil
	}

	return nil, fmt.Errorf("[Unknown] invalid packet type %d", t)
}

// Valid returns a boolean indicating whether the type is valid or not.
func (t Type) Valid() bool {
	return t >= CONNECT && t <= DISCONNECT
}
