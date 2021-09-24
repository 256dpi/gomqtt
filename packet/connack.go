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
	return cc <= 5
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

	// If a well-formed Connect packet is received by the server, but the server
	// is unable to process it for some reason, then the server should attempt
	// to send a Connack containing a non-zero ReturnCode.
	ReturnCode ConnackCode
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
	return fmt.Sprintf("<Connack SessionPresent=%t ReturnCode=%d>",
		c.SessionPresent, c.ReturnCode)
}

// Len returns the byte length of the encoded packet.
func (c *Connack) Len() int {
	return headerLen(2) + 2
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (c *Connack) Decode(src []byte) (int, error) {
	// decode header
	total, _, _, err := decodeHeader(src, CONNACK)
	if err != nil {
		return total, err
	}

	// read connack flags
	connackFlags, n, err := readUint8(src[total:], CONNACK)
	total += n
	if err != nil {
		return total, err
	}

	// check flags
	if connackFlags&254 != 0 {
		return total, makeError(CONNACK, "bits 7-1 in acknowledge flags are not 0")
	}

	// set session present
	c.SessionPresent = connackFlags&0x1 == 1

	// read return code
	rc, n, err := readUint8(src[total:], CONNACK)
	total += n
	if err != nil {
		return total, err
	}

	// get return code
	returnCode := ConnackCode(rc)
	if !returnCode.Valid() {
		return total, makeError(CONNACK, "invalid return code (%d)", c.ReturnCode)
	}

	// set return code
	c.ReturnCode = returnCode

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (c *Connack) Encode(dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, 2, c.Len(), CONNACK)
	if err != nil {
		return total, err
	}

	// get connack flags
	var flags uint8
	if c.SessionPresent {
		flags = 0x1 // 00000001
	} else {
		flags = 0x0 // 00000000
	}

	// write flags
	n, err := writeUint8(dst[total:], flags, CONNACK)
	total += n
	if err != nil {
		return total, err
	}

	// check return code
	if !c.ReturnCode.Valid() {
		return total, makeError(CONNACK, "invalid return code (%d)", c.ReturnCode)
	}

	// write return code
	n, err = writeUint8(dst[total:], uint8(c.ReturnCode), CONNACK)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}
