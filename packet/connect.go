package packet

import (
	"bytes"
	"fmt"
)

// The supported MQTT versions.
const (
	Version3 byte = 0x3
	Version4 byte = 0x4
	Version5 byte = 0x5
)

var versionNames = map[byte][]byte{
	Version3: []byte("MQIsdp"),
	Version4: []byte("MQTT"),
	Version5: []byte("MQTT"),
}

// A Connect packet is sent by a client to the server after a network
// connection has been established.
type Connect struct {
	// The clients client id.
	ClientID string

	// The keep alive value.
	KeepAlive uint16

	// The authentication username.
	Username string

	// The authentication password.
	Password string

	// The clean session flag.
	CleanSession bool

	// The will message.
	Will *Message

	// The MQTT version.
	Version byte

	CleanStart bool

	WillProperties []Property
	// WillDelayInterval uint32
	// WillPayloadFormatIndicator byte
	// WillMessageExpiryInterval uint32
	// WillContentType string
	// WillResponseTopic string
	// WillCorrelationData []byte
	// WillUserProperties map[string][]byte

	Properties []Property
	// SessionExpiryInterval uint32
	// ReceiveMaximum uint16
	// MaximumPacketSize uint32
	// TopicAliasMaximum uint16
	// RequestResponseInformation bool
	// RequestProblemInformation bool
	// AuthenticationMethod string
	// AuthenticationData []byte
	// UserProperties map[string][]byte
}

// NewConnect creates a new Connect packet.
func NewConnect() *Connect {
	return &Connect{
		CleanSession: true,
		Version:      Version4,
	}
}

// Type returns the packets type.
func (c *Connect) Type() Type {
	return CONNECT
}

// String returns a string representation of the packet.
func (c *Connect) String() string {
	// prepare will
	will := "nil"
	if c.Will != nil {
		will = c.Will.String()
	}

	return fmt.Sprintf(
		"<Connect ClientID=%q KeepAlive=%d Username=%q Password=%q CleanSession=%t Will=%s Version=%d>",
		c.ClientID, c.KeepAlive, c.Username, c.Password, c.CleanSession, will, c.Version,
	)
}

// Len returns the byte length of the encoded packet.
func (c *Connect) Len(m Mode) int {
	ml := c.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (c *Connect) Decode(m Mode, src []byte) (int, error) {
	// decode header
	total, _, _, err := decodeHeader(src, CONNECT)
	if err != nil {
		return total, wrapError(CONNECT, DECODE, m, total, err)
	}

	// read protocol string
	protoName, n, err := readBytes(src[total:], false)
	total += n
	if err != nil {
		return total, wrapError(CONNECT, DECODE, m, total, err)
	}

	// read version
	versionByte, n, err := readUint8(src[total:])
	total += n
	if err != nil {
		return total, wrapError(CONNECT, DECODE, m, total, err)
	}

	// check protocol version
	if versionByte < Version3 || versionByte > Version5 {
		return total, makeError(CONNECT, DECODE, m, total, "invalid protocol version")
	}

	// set version
	c.Version = versionByte

	// check protocol version string
	if !bytes.Equal(protoName, versionNames[c.Version]) {
		return total, makeError(CONNECT, DECODE, m, total, "invalid protocol string")
	}

	// get connect flags
	connectFlags, n, err := readUint8(src[total:])
	total += n
	if err != nil {
		return total, wrapError(CONNECT, DECODE, m, total, err)
	}

	// read flags
	usernameFlag := ((connectFlags >> 7) & 0b1) == 1
	passwordFlag := ((connectFlags >> 6) & 0b1) == 1
	willFlag := ((connectFlags >> 2) & 0b1) == 1
	willRetain := ((connectFlags >> 5) & 0b1) == 1
	willQOS := QOS((connectFlags >> 3) & 0b11)
	c.CleanSession = ((connectFlags >> 1) & 0b1) == 1

	// check reserved bit
	if connectFlags&0x1 != 0 {
		return total, makeError(CONNECT, DECODE, m, total, "invalid connect flags")
	}

	// check will qos
	if !willQOS.Successful() {
		return total, wrapError(CONNECT, DECODE, m, total, ErrInvalidQOSLevel)
	}

	// check will flags
	if !willFlag && (willRetain || willQOS != 0) {
		return total, makeError(CONNECT, DECODE, m, total, "invalid connect flags")
	}

	// create will if present
	if willFlag {
		c.Will = &Message{QOS: willQOS, Retain: willRetain}
	}

	// check auth flags
	if !usernameFlag && passwordFlag {
		return total, makeError(CONNECT, DECODE, m, total, "invalid connect flags")
	}

	// read keep alive
	ka, n, err := readUint(src[total:], 2)
	total += n
	if err != nil {
		return total, wrapError(CONNECT, DECODE, m, total, err)
	}

	// set keep alive
	c.KeepAlive = uint16(ka)

	// read client id
	c.ClientID, n, err = readString(src[total:])
	total += n
	if err != nil {
		return total, wrapError(CONNECT, DECODE, m, total, err)
	}

	// if the client supplies a zero-byte clientID, the client must also set CleanSession to 1
	if len(c.ClientID) == 0 && !c.CleanSession {
		return total, makeError(CONNECT, DECODE, m, total, "missing client id")
	}

	// check will
	if c.Will != nil {
		// read will topic
		c.Will.Topic, n, err = readString(src[total:])
		total += n
		if err != nil {
			return total, wrapError(CONNECT, DECODE, m, total, err)
		}

		// read will payload
		c.Will.Payload, n, err = readBytes(src[total:], true)
		total += n
		if err != nil {
			return total, wrapError(CONNECT, DECODE, m, total, err)
		}
	}

	// read username
	if usernameFlag {
		c.Username, n, err = readString(src[total:])
		total += n
		if err != nil {
			return total, wrapError(CONNECT, DECODE, m, total, err)
		}
	}

	// read password
	if passwordFlag {
		c.Password, n, err = readString(src[total:])
		total += n
		if err != nil {
			return total, wrapError(CONNECT, DECODE, m, total, err)
		}
	}

	// check buffer
	if len(src[total:]) != 0 {
		return total, makeError(CONNECT, DECODE, m, total, "leftover buffer length")
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (c *Connect) Encode(m Mode, dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, c.len(), CONNECT)
	if err != nil {
		return total, wrapError(CONNECT, ENCODE, m, total, err)
	}

	// set default version byte
	if c.Version == 0 {
		c.Version = m.Version
	}

	// check version byte
	if c.Version < Version3 || c.Version > Version5 {
		return total, makeError(CONNECT, ENCODE, m, total, "invalid protocol version")
	}

	// write version string
	n, err := writeBytes(dst[total:], versionNames[c.Version])
	total += n
	if err != nil {
		return total, wrapError(CONNECT, ENCODE, m, total, err)
	}

	// write version value
	n, err = writeUint8(dst[total:], c.Version)
	total += n
	if err != nil {
		return total, wrapError(CONNECT, ENCODE, m, total, err)
	}

	// prepare connect flags
	var connectFlags byte

	// set username flag
	if len(c.Username) > 0 {
		connectFlags |= 0b1000_0000
	}

	// set password flag
	if len(c.Password) > 0 {
		connectFlags |= 0b100_0000
	}

	// check will
	if c.Will != nil {
		// set will flag
		connectFlags |= 0b100

		// check will topic length
		if len(c.Will.Topic) == 0 {
			return total, makeError(CONNECT, ENCODE, m, total, "missing will topic")
		}

		// check will qos
		if !c.Will.QOS.Successful() {
			return total, wrapError(CONNECT, ENCODE, m, total, ErrInvalidQOSLevel)
		}

		// set will qos flag
		connectFlags = (connectFlags & 0b1110_0111) | (byte(c.Will.QOS) << 3)

		// set will retain flag
		if c.Will.Retain {
			connectFlags |= 0b010_0000
		}
	}

	// check client id and clean session
	if len(c.ClientID) == 0 && !c.CleanSession {
		return total, makeError(CONNECT, ENCODE, m, total, "missing client id")
	}

	// set clean session flag
	if c.CleanSession {
		connectFlags |= 0b10
	}

	// write connect flags
	n, err = writeUint8(dst[total:], connectFlags)
	total += n
	if err != nil {
		return total, wrapError(CONNECT, ENCODE, m, total, err)
	}

	// write keep alive
	n, err = writeUint(dst[total:], uint64(c.KeepAlive), 2)
	total += n
	if err != nil {
		return total, wrapError(CONNECT, ENCODE, m, total, err)
	}

	// write client id
	n, err = writeString(dst[total:], c.ClientID)
	total += n
	if err != nil {
		return total, wrapError(CONNECT, ENCODE, m, total, err)
	}

	// check will
	if c.Will != nil {
		// write will topic
		n, err = writeString(dst[total:], c.Will.Topic)
		total += n
		if err != nil {
			return total, wrapError(CONNECT, ENCODE, m, total, err)
		}

		// write will payload
		n, err = writeBytes(dst[total:], c.Will.Payload)
		total += n
		if err != nil {
			return total, wrapError(CONNECT, ENCODE, m, total, err)
		}
	}

	// check username and password
	if len(c.Username) == 0 && len(c.Password) > 0 {
		return total, makeError(CONNECT, ENCODE, m, total, "missing username")
	}

	// write username
	if len(c.Username) > 0 {
		n, err = writeString(dst[total:], c.Username)
		total += n
		if err != nil {
			return total, wrapError(CONNECT, ENCODE, m, total, err)
		}
	}

	// write password
	if len(c.Password) > 0 {
		n, err = writeString(dst[total:], c.Password)
		total += n
		if err != nil {
			return total, wrapError(CONNECT, ENCODE, m, total, err)
		}
	}

	return total, nil
}

func (c *Connect) len() int {
	// prepare total
	total := 0

	// add protocol string and version
	total += 2 + len(versionNames[c.Version]) + 1

	// add 1 byte connect flags
	// add 2 bytes keep alive timer
	total += 1 + 2

	// add the clientID length
	total += 2 + len(c.ClientID)

	// add the will topic and will message length
	if c.Will != nil {
		total += 2 + len(c.Will.Topic) + 2 + len(c.Will.Payload)
	}

	// add the username length
	if len(c.Username) > 0 {
		total += 2 + len(c.Username)
	}

	// add the password length
	if len(c.Password) > 0 {
		total += 2 + len(c.Password)
	}

	return total
}
