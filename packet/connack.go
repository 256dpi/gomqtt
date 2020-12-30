package packet

import "fmt"

// The ConnackCode represents the return code in a Connack packet.
type ConnackCode uint8

// All available ConnackCodes.
const (
	ConnectionAccepted ConnackCode = iota
	InvalidProtocolVersion
	IdentifierRejected
	ServerUnavailable
	BadUsernameOrPassword
	NotAuthorized
)

// Valid checks if the ConnackCode is valid.
func (cc ConnackCode) Valid() bool {
	return cc <= NotAuthorized
}

// String returns the corresponding error string for the ConnackCode.
func (cc ConnackCode) String() string {
	switch cc {
	case ConnectionAccepted:
		return "connection accepted"
	case InvalidProtocolVersion:
		return "connection refused: unacceptable protocol version"
	case IdentifierRejected:
		return "connection refused: identifier rejected"
	case ServerUnavailable:
		return "connection refused: server unavailable"
	case BadUsernameOrPassword:
		return "connection refused: bad user name or password"
	case NotAuthorized:
		return "connection refused: not authorized"
	}

	return "invalid connack code"
}

// A Connack packet is sent by the server in response to a Connect packet
// received from a client.
type Connack struct {
	// The SessionPresent flag enables a client to establish whether the
	// client and server have a consistent view about whether there is already
	// stored session state.
	SessionPresent bool

	// If a well formed Connect packet is received by the server, but the server
	// is unable to process it for some reason, then the server should attempt
	// to send a Connack containing a non-zero ReturnCode.
	ReturnCode ConnackCode

	ReasonCode byte

	Properties []Property
	// SessionExpiryInterval uint32
	// ReceiveMaximum uint16
	// MaximumQOS byte
	// RetainAvailable bool
	// MaximumPacketSize uint32
	// AssignedClientID string
	// TopicAliasMaximum uint16
	// ReasonString string
	// UserProperties map[string][]byte
	// WildcardSubscriptionAvailable bool
	// SubscriptionIdentifiersAvailable bool
	// SharedSubscriptionAvailable bool
	// ServerKeepAlive uint16
	// ResponseInformation string
	// ServerReference string
}

// NewConnack creates a new Connack packet.
func NewConnack() *Connack {
	return &Connack{}
}

// Type returns the packets type.
func (c *Connack) Type() Type {
	return CONNACK
}

// String returns a string representation of the packet.
func (c *Connack) String() string {
	return fmt.Sprintf(
		"<Connack SessionPresent=%t ReturnCode=%d>",
		c.SessionPresent, c.ReturnCode,
	)
}

// Len returns the byte length of the encoded packet.
func (c *Connack) Len(m Mode) int {
	return headerLen(2) + 2
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (c *Connack) Decode(m Mode, src []byte) (int, error) {
	// decode header
	total, _, _, err := decodeHeader(src, CONNACK)
	if err != nil {
		return total, wrapError(CONNACK, DECODE, m, total, err)
	}

	// read connack flags
	connackFlags, n, err := readUint8(src[total:])
	total += n
	if err != nil {
		return total, wrapError(CONNACK, DECODE, m, total, err)
	}

	// check flags
	if connackFlags&0b1111_1110 != 0 {
		return total, makeError(CONNACK, DECODE, m, total, "invalid connack flags")
	}

	// set session present
	c.SessionPresent = connackFlags&0b1 == 1

	// read return code
	rc, n, err := readUint8(src[total:])
	total += n
	if err != nil {
		return total, wrapError(CONNACK, DECODE, m, total, err)
	}

	// get return code
	returnCode := ConnackCode(rc)
	if !returnCode.Valid() {
		return total, makeError(CONNACK, DECODE, m, total, "invalid return code")
	}

	// set return code
	c.ReturnCode = returnCode

	// check buffer
	if len(src[total:]) != 0 {
		return total, makeError(CONNACK, DECODE, m, total, "leftover buffer length")
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (c *Connack) Encode(m Mode, dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, 2, CONNACK)
	if err != nil {
		return total, wrapError(CONNACK, ENCODE, m, total, err)
	}

	// get connack flags
	var flags uint8
	if c.SessionPresent {
		flags = 0b1
	} else {
		flags = 0b0
	}

	// write flags
	n, err := writeUint8(dst[total:], flags)
	total += n
	if err != nil {
		return total, wrapError(CONNACK, ENCODE, m, total, err)
	}

	// check return code
	if !c.ReturnCode.Valid() {
		return total, makeError(CONNACK, ENCODE, m, total, "invalid return code")
	}

	// write return code
	n, err = writeUint8(dst[total:], uint8(c.ReturnCode))
	total += n
	if err != nil {
		return total, wrapError(CONNACK, ENCODE, m, total, err)
	}

	return total, nil
}
