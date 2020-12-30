package packet

import "errors"

// ErrInvalidPacketType is returned by New if the packet type is invalid.
var ErrInvalidPacketType = errors.New("invalid packet type")

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
	AUTH
)

// Types returns a list of all known packet types.
func Types() []Type {
	return []Type{CONNECT, CONNACK, PUBLISH, PUBACK, PUBREC, PUBREL, PUBCOMP,
		SUBSCRIBE, SUBACK, UNSUBSCRIBE, UNSUBACK, PINGREQ, PINGRESP, DISCONNECT,
		AUTH}
}

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
	case AUTH:
		return "Auth"
	}

	return "Unknown"
}

func (t Type) defaultFlags() byte {
	switch t {
	case PUBREL:
		return 0b10
	case SUBSCRIBE:
		return 0b10
	case UNSUBSCRIBE:
		return 0b10
	}

	return 0
}

// New creates a new packet based on the type. It is a shortcut to call one of
// the New*Packet functions. An error is returned if the type is invalid.
func (t Type) New() (Generic, error) {
	switch t {
	case CONNECT:
		return NewConnect(), nil
	case CONNACK:
		return NewConnack(), nil
	case PUBLISH:
		return NewPublish(), nil
	case PUBACK:
		return NewPuback(), nil
	case PUBREC:
		return NewPubrec(), nil
	case PUBREL:
		return NewPubrel(), nil
	case PUBCOMP:
		return NewPubcomp(), nil
	case SUBSCRIBE:
		return NewSubscribe(), nil
	case SUBACK:
		return NewSuback(), nil
	case UNSUBSCRIBE:
		return NewUnsubscribe(), nil
	case UNSUBACK:
		return NewUnsuback(), nil
	case PINGREQ:
		return NewPingreq(), nil
	case PINGRESP:
		return NewPingresp(), nil
	case DISCONNECT:
		return NewDisconnect(), nil
	case AUTH:
		return NewAuth(), nil
	}

	return nil, ErrInvalidPacketType
}

// Valid returns a boolean indicating whether the type is valid or not.
func (t Type) Valid() bool {
	return t >= CONNECT && t <= AUTH
}
