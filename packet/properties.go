package packet

import (
	"encoding/binary"
	"errors"
)

// ErrInvalidPropertyType is returned if the property type is not known.
var ErrInvalidPropertyType = errors.New("invalid property type")

// ErrInvalidPropertyCode is returned if the property code is not known.
var ErrInvalidPropertyCode = errors.New("invalid property code")

// PropertyType defines the properties type.
type PropertyType int

// The available property types.
const (
	UINT8 PropertyType = iota + 1
	UINT16
	UINT32
	VARINT
	STRING
	BYTES
	PAIR
)

// String returns the property type name.
func (t PropertyType) String() string {
	switch t {
	case UINT8:
		return "UINT8"
	case UINT16:
		return "UINT16"
	case UINT32:
		return "UINT32"
	case VARINT:
		return "VARINT"
	case STRING:
		return "STRING"
	case BYTES:
		return "BYTES"
	case PAIR:
		return "PAIR"
	default:
		return ""
	}
}

// Valid returns whether the property type is valid.
func (t PropertyType) Valid() bool {
	return t >= UINT8 && t <= PAIR
}

// PropertyCode defines a property code.
type PropertyCode byte

// The available property codes.
const (
	PayloadFormatIndicator          PropertyCode = 0x01
	MessageExpiryInterval           PropertyCode = 0x02
	ContentType                     PropertyCode = 0x03
	ResponseTopic                   PropertyCode = 0x08
	CorrelationData                 PropertyCode = 0x09
	SubscriptionIdentifier          PropertyCode = 0x0B
	SessionExpiryInterval           PropertyCode = 0x11
	AssignedClientIdentifier        PropertyCode = 0x12
	ServerKeepAlive                 PropertyCode = 0x13
	AuthenticationMethod            PropertyCode = 0x15
	AuthenticationData              PropertyCode = 0x16
	RequestProblemInformation       PropertyCode = 0x17
	WillDelayInterval               PropertyCode = 0x18
	RequestResponseInformation      PropertyCode = 0x19
	ResponseInformation             PropertyCode = 0x1A
	ServerReference                 PropertyCode = 0x1C
	ReasonString                    PropertyCode = 0x1F
	ReceiveMaximum                  PropertyCode = 0x21
	TopicAliasMaximum               PropertyCode = 0x22
	TopicAlias                      PropertyCode = 0x23
	MaximumQOS                      PropertyCode = 0x24
	RetainAvailable                 PropertyCode = 0x25
	UserProperty                    PropertyCode = 0x26
	MaximumPacketSize               PropertyCode = 0x27
	WildcardSubscriptionAvailable   PropertyCode = 0x28
	SubscriptionIdentifierAvailable PropertyCode = 0x29
	SharedSubscriptionAvailable     PropertyCode = 0x2A
)

var propertyTypes = map[PropertyCode]PropertyType{
	PayloadFormatIndicator:          UINT8,
	MessageExpiryInterval:           UINT32,
	ContentType:                     STRING,
	ResponseTopic:                   STRING,
	CorrelationData:                 BYTES,
	SubscriptionIdentifier:          VARINT,
	SessionExpiryInterval:           UINT32,
	AssignedClientIdentifier:        STRING,
	ServerKeepAlive:                 UINT16,
	AuthenticationMethod:            STRING,
	AuthenticationData:              BYTES,
	RequestProblemInformation:       UINT8,
	WillDelayInterval:               UINT32,
	RequestResponseInformation:      UINT8,
	ResponseInformation:             STRING,
	ServerReference:                 STRING,
	ReasonString:                    STRING,
	ReceiveMaximum:                  UINT16,
	TopicAliasMaximum:               UINT16,
	TopicAlias:                      UINT16,
	MaximumQOS:                      UINT8,
	RetainAvailable:                 UINT8,
	UserProperty:                    PAIR,
	MaximumPacketSize:               STRING,
	WildcardSubscriptionAvailable:   UINT8,
	SubscriptionIdentifierAvailable: UINT8,
	SharedSubscriptionAvailable:     UINT8,
}

var propertyNames = map[PropertyCode]string{
	PayloadFormatIndicator:          "PayloadFormatIndicator",
	MessageExpiryInterval:           "MessageExpiryInterval",
	ContentType:                     "ContentType",
	ResponseTopic:                   "ResponseTopic",
	CorrelationData:                 "CorrelationData",
	SubscriptionIdentifier:          "SubscriptionIdentifier",
	SessionExpiryInterval:           "SessionExpiryInterval",
	AssignedClientIdentifier:        "AssignedClientIdentifier",
	ServerKeepAlive:                 "ServerKeepAlive",
	AuthenticationMethod:            "AuthenticationMethod",
	AuthenticationData:              "AuthenticationData",
	RequestProblemInformation:       "RequestProblemInformation",
	WillDelayInterval:               "WillDelayInterval",
	RequestResponseInformation:      "RequestResponseInformation",
	ResponseInformation:             "ResponseInformation",
	ServerReference:                 "ServerReference",
	ReasonString:                    "ReasonString",
	ReceiveMaximum:                  "ReceiveMaximum",
	TopicAliasMaximum:               "TopicAliasMaximum",
	TopicAlias:                      "TopicAlias",
	MaximumQOS:                      "MaximumQOS",
	RetainAvailable:                 "RetainAvailable",
	UserProperty:                    "UserProperty",
	MaximumPacketSize:               "MaximumPacketSize",
	WildcardSubscriptionAvailable:   "WildcardSubscriptionAvailable",
	SubscriptionIdentifierAvailable: "SubscriptionIdentifierAvailable",
	SharedSubscriptionAvailable:     "SharedSubscriptionAvailable",
}

// Type returns the property type.
func (c PropertyCode) Type() PropertyType {
	return propertyTypes[c]
}

// String returns the property code name.
func (c PropertyCode) String() string {
	return propertyNames[c]
}

// Valid returns whether the property code is known.
func (c PropertyCode) Valid() bool {
	return c.Type().Valid() && c.String() != ""
}

// Property is a MQTT packet property.
type Property struct {
	// The property code.
	Code PropertyCode

	// The UINT8, UINT16, UINT32 or VARINT value.
	Uint64 uint64

	// The STRING value.
	String string

	// The BYTES value.
	Bytes []byte

	// The PAIR key and value.
	Key, Value string
}

func (p *Property) len() int {
	// get code len
	cl := varintLen(uint64(p.Code))

	// get value len
	vl := 0
	switch p.Code.Type() {
	case UINT8:
		vl = 1
	case UINT16:
		vl = 2
	case UINT32:
		vl = 4
	case VARINT:
		vl = varintLen(p.Uint64)
	case STRING:
		vl = 2 + len(p.String)
	case PAIR:
		vl = 4 + len(p.Key) + len(p.Value)
	case BYTES:
		vl = 2 + len(p.Bytes)
	}

	return cl + vl
}

func (p *Property) decode(buf []byte) (int, error) {
	// read value
	var n int
	var err error
	switch p.Code.Type() {
	case UINT8:
		p.Uint64, n, err = readUint(buf, 1)
	case UINT16:
		p.Uint64, n, err = readUint(buf, 2)
	case UINT32:
		p.Uint64, n, err = readUint(buf, 4)
	case VARINT:
		p.Uint64, n, err = readVarint(buf)
	case STRING:
		p.String, n, err = readString(buf, nil)
	case BYTES:
		p.Bytes, n, err = readBytes(buf, true)
	case PAIR:
		p.Key, p.Value, n, err = readPair(buf)
	default:
		return 0, ErrInvalidPropertyType
	}

	return n, err
}

func (p *Property) encode(buf []byte) (int, error) {
	// write value
	switch p.Code.Type() {
	case UINT8:
		return writeUint(buf, p.Uint64, 1)
	case UINT16:
		return writeUint(buf, p.Uint64, 2)
	case UINT32:
		return writeUint(buf, p.Uint64, 4)
	case VARINT:
		return writeVarint(buf, p.Uint64)
	case STRING:
		return writeString(buf, p.String)
	case BYTES:
		return writeBytes(buf, p.Bytes)
	case PAIR:
		return writePair(buf, p.Key, p.Value)
	default:
		return 0, ErrInvalidPropertyType
	}
}

func propertiesLen(properties []Property) int {
	// get properties length length
	pll := varintLen(uint64(len(properties)))

	// sum properties
	sum := 0
	for _, property := range properties {
		sum += property.len()
	}

	return pll + sum
}

func readProperties(buf []byte) ([]Property, int, error) {
	// prepare read counter
	read := 0

	// read property length
	pl, n, err := readVarint(buf)
	read += n
	if err != nil {
		return nil, read, err
	}

	// prepare properties
	properties := make([]Property, int(pl))

	// read properties
	for i := 0; i < int(pl); i++ {
		// read property code
		pc, n := binary.Uvarint(buf[read:])
		read += n
		if n <= 0 {
			return nil, read, ErrInvalidPropertyType
		}

		// get code
		code := PropertyCode(pc)

		// check code
		if !code.Valid() {
			return nil, read, ErrInvalidPropertyCode
		}

		// decode property
		n, err := properties[i].decode(buf[read:])
		read += n
		if err != nil {
			return nil, read, err
		}
	}

	return properties, read, nil
}

func writeProperties(buf []byte, properties []Property) (int, error) {
	// prepare write counter
	write := 0

	// write property length
	n, err := writeVarint(buf, uint64(len(properties)))
	write += n
	if err != nil {
		return write, err
	}

	// write properties
	for _, property := range properties {
		// check code
		if !property.Code.Valid() {
			return write, ErrInvalidPropertyCode
		}

		// encode property
		n, err := property.encode(buf[write:])
		write += n
		if err != nil {
			return write, err
		}
	}

	return write, nil
}
