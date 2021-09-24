package packet

import (
	"bytes"
	"fmt"
)

// The supported MQTT versions.
const (
	Version31  byte = 3
	Version311 byte = 4
)

var versionNames = map[byte][]byte{
	Version31:  []byte("MQIsdp"),
	Version311: []byte("MQTT"),
}

// A Connect packet is sent by a client to the server after a network
// connection has been established.
type Connect struct {
	// The client's client id.
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

	// The MQTT version 3 or 4 (defaults to 4 when 0).
	Version byte
}

// NewConnect creates a new Connect packet.
func NewConnect() *Connect {
	return &Connect{
		CleanSession: true,
		Version:      4,
	}
}

// Type returns the packets type.
func (c *Connect) Type() Type {
	return CONNECT
}

// String returns a string representation of the packet.
func (c *Connect) String() string {
	will := "nil"

	if c.Will != nil {
		will = c.Will.String()
	}

	return fmt.Sprintf("<Connect ClientID=%q KeepAlive=%d Username=%q "+
		"Password=%q CleanSession=%t Will=%s Version=%d>",
		c.ClientID,
		c.KeepAlive,
		c.Username,
		c.Password,
		c.CleanSession,
		will,
		c.Version,
	)
}

// Len returns the byte length of the encoded packet.
func (c *Connect) Len() int {
	ml := c.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (c *Connect) Decode(src []byte) (int, error) {
	// decode header
	total, _, _, err := decodeHeader(src, CONNECT)
	if err != nil {
		return total, err
	}

	// read protocol string
	protoName, n, err := readLPBytes(src[total:], false, CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// read version
	versionByte, n, err := readUint8(src[total:], CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// check protocol version
	if versionByte != Version311 && versionByte != Version31 {
		return total, makeError(CONNECT, "invalid protocol version (%d)", versionByte)
	}

	// set version
	c.Version = versionByte

	// check protocol version string
	if !bytes.Equal(protoName, versionNames[c.Version]) {
		return total, makeError(CONNECT, "invalid protocol version string (%s)", protoName)
	}

	// get connect flags
	connectFlags, n, err := readUint8(src[total:], CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// read flags
	usernameFlag := ((connectFlags >> 7) & 0x1) == 1
	passwordFlag := ((connectFlags >> 6) & 0x1) == 1
	willFlag := ((connectFlags >> 2) & 0x1) == 1
	willRetain := ((connectFlags >> 5) & 0x1) == 1
	willQOS := QOS((connectFlags >> 3) & 0x3)
	c.CleanSession = ((connectFlags >> 1) & 0x1) == 1

	// check reserved bit
	if connectFlags&0x1 != 0 {
		return total, makeError(CONNECT, "reserved bit 0 is not 0")
	}

	// check will qos
	if !willQOS.Successful() {
		return total, makeError(CONNECT, "invalid QOS level (%d) for will message", willQOS)
	}

	// check will flags
	if !willFlag && (willRetain || willQOS != 0) {
		return total, makeError(CONNECT, "if the will flag is set to 0 the will qos and will retain fields must be set to zero")
	}

	// create will if present
	if willFlag {
		c.Will = &Message{QOS: willQOS, Retain: willRetain}
	}

	// check auth flags
	if !usernameFlag && passwordFlag {
		return total, makeError(CONNECT, "password flag is set but username flag is not set")
	}

	// read keep alive
	ka, n, err := readUint(src[total:], 2, CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// set keep alive
	c.KeepAlive = uint16(ka)

	// read client id
	c.ClientID, n, err = readLPString(src[total:], CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// if the client supplies a zero-byte clientID, the client must also set CleanSession to 1
	if len(c.ClientID) == 0 && !c.CleanSession {
		return total, makeError(CONNECT, "clean session must be 1 if client id is zero length")
	}

	// check will
	if c.Will != nil {
		// read will topic
		c.Will.Topic, n, err = readLPString(src[total:], CONNECT)
		total += n
		if err != nil {
			return total, err
		}

		// read will payload
		c.Will.Payload, n, err = readLPBytes(src[total:], true, CONNECT)
		total += n
		if err != nil {
			return total, err
		}
	}

	// read username
	if usernameFlag {
		c.Username, n, err = readLPString(src[total:], CONNECT)
		total += n
		if err != nil {
			return total, err
		}
	}

	// read password
	if passwordFlag {
		c.Password, n, err = readLPString(src[total:], CONNECT)
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
func (c *Connect) Encode(dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, c.len(), c.Len(), CONNECT)
	if err != nil {
		return total, err
	}

	// set default version byte
	if c.Version == 0 {
		c.Version = Version311
	}

	// check version byte
	if c.Version != Version311 && c.Version != Version31 {
		return total, makeError(CONNECT, "unsupported protocol version %d", c.Version)
	}

	// write version string
	n, err := writeLPBytes(dst[total:], versionNames[c.Version], CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// write version value
	n, err = writeUint8(dst[total:], c.Version, CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// prepare connect flags
	var connectFlags byte

	// set username flag
	if len(c.Username) > 0 {
		connectFlags |= 128 // 10000000
	}

	// set password flag
	if len(c.Password) > 0 {
		connectFlags |= 64 // 01000000
	}

	// check will
	if c.Will != nil {
		// set will flag
		connectFlags |= 0x4 // 00000100

		// check will topic length
		if len(c.Will.Topic) == 0 {
			return total, makeError(CONNECT, "will topic is empty")
		}

		// check will qos
		if !c.Will.QOS.Successful() {
			return total, makeError(CONNECT, "invalid will qos level %d", c.Will.QOS)
		}

		// set will qos flag
		connectFlags = (connectFlags & 231) | (byte(c.Will.QOS) << 3) // 231 = 11100111

		// set will retain flag
		if c.Will.Retain {
			connectFlags |= 32 // 00100000
		}
	}

	// check client id and clean session
	if len(c.ClientID) == 0 && !c.CleanSession {
		return total, makeError(CONNECT, "clean session must be 1 if client id is zero length")
	}

	// set clean session flag
	if c.CleanSession {
		connectFlags |= 0x2 // 00000010
	}

	// write connect flags
	n, err = writeUint8(dst[total:], connectFlags, CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// write keep alive
	n, err = writeUint(dst[total:], uint64(c.KeepAlive), 2, CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// write client id
	n, err = writeLPString(dst[total:], c.ClientID, CONNECT)
	total += n
	if err != nil {
		return total, err
	}

	// check will
	if c.Will != nil {
		// write will topic
		n, err = writeLPString(dst[total:], c.Will.Topic, CONNECT)
		total += n
		if err != nil {
			return total, err
		}

		// write will payload
		n, err = writeLPBytes(dst[total:], c.Will.Payload, CONNECT)
		total += n
		if err != nil {
			return total, err
		}
	}

	// check username and password
	if len(c.Username) == 0 && len(c.Password) > 0 {
		return total, makeError(CONNECT, "password set without username")
	}

	// write username
	if len(c.Username) > 0 {
		n, err = writeLPString(dst[total:], c.Username, CONNECT)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write password
	if len(c.Password) > 0 {
		n, err = writeLPString(dst[total:], c.Password, CONNECT)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (c *Connect) len() int {
	// prepare total
	total := 0

	// add version
	if c.Version == Version31 {
		// 2 bytes protocol name length
		// 6 bytes protocol name
		// 1 byte protocol version
		total += 2 + 6 + 1
	} else {
		// 2 bytes protocol name length
		// 4 bytes protocol name
		// 1 byte protocol version
		total += 2 + 4 + 1
	}

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
