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

// Type represents the MQTT packet types.
type Type byte

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
		return "CONNECT"
	case CONNACK:
		return "CONNACK"
	case PUBLISH:
		return "PUBLISH"
	case PUBACK:
		return "PUBACK"
	case PUBREC:
		return "PUBREC"
	case PUBREL:
		return "PUBREL"
	case PUBCOMP:
		return "PUBCOMP"
	case SUBSCRIBE:
		return "SUBSCRIBE"
	case SUBACK:
		return "SUBACK"
	case UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case UNSUBACK:
		return "UNSUBACK"
	case PINGREQ:
		return "PINGREQ"
	case PINGRESP:
		return "PINGRESP"
	case DISCONNECT:
		return "DISCONNECT"
	}

	return "UNKNOWN"
}

// DefaultFlags returns the default flag values for the message type, as defined by
// the MQTT spec, except for PUBLISH.
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

// New creates a new message based on the type. It is a shortcut to call one of the
// New*Message functions. If an error is returned then the type is invalid.
func (t Type) New() (Message, error) {
	switch t {
	case CONNECT:
		return NewConnectMessage(), nil
	case CONNACK:
		return NewConnackMessage(), nil
	case PUBLISH:
		return NewPublishMessage(), nil
	case PUBACK:
		return NewPubackMessage(), nil
	case PUBREC:
		return NewPubrecMessage(), nil
	case PUBREL:
		return NewPubrelMessage(), nil
	case PUBCOMP:
		return NewPubcompMessage(), nil
	case SUBSCRIBE:
		return NewSubscribeMessage(), nil
	case SUBACK:
		return NewSubackMessage(), nil
	case UNSUBSCRIBE:
		return NewUnsubscribeMessage(), nil
	case UNSUBACK:
		return NewUnsubackMessage(), nil
	case PINGREQ:
		return NewPingreqMessage(), nil
	case PINGRESP:
		return NewPingrespMessage(), nil
	case DISCONNECT:
		return NewDisconnectMessage(), nil
	}

	return nil, fmt.Errorf("Invalid type %d", t)
}

// Valid returns a boolean indicating whether the type is valid or not.
func (t Type) Valid() bool {
	return t >= CONNECT && t <= DISCONNECT
}
