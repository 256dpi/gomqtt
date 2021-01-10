package packet

import (
	"fmt"
)

// returns the byte length of an identified packet
func identifiedLen() int {
	return headerLen(2) + 2
}

// decodes an identified packet
func identifiedDecode(m Mode, src []byte, id *ID, t Type) (int, error) {
	// decode header
	total, _, _, err := decodeHeader(src, t)
	if err != nil {
		return total, err
	}

	// read packet id
	pid, n, err := readUint(src[total:], 2, t)
	total += n
	if err != nil {
		return total, err
	}

	// get packet id
	packetID := ID(pid)
	if !packetID.Valid() {
		return total, makeError(t, "packet id must be grater than zero")
	}

	// set packet id
	*id = packetID

	// check buffer
	if len(src[total:]) != 0 {
		return total, makeError(CONNACK, "leftover buffer length (%d)", len(src[total:]))
	}

	return total, nil
}

// encodes an identified packet
func identifiedEncode(m Mode, dst []byte, id ID, t Type) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, 2, t)
	if err != nil {
		return total, err
	}

	// check packet id
	if !id.Valid() {
		return total, makeError(t, "packet id must be grater than zero")
	}

	// write packet id
	n, err := writeUint(dst[total:], uint64(id), 2, t)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

// A Puback packet is the response to a Publish packet with QOS level 1.
type Puback struct {
	// The packet identifier.
	ID ID

	ReasonCode byte

	Properties []Property
	// ReasonString string
	// UserProperties map[string][]byte
}

// NewPuback creates a new Puback packet.
func NewPuback() *Puback {
	return &Puback{}
}

// Type returns the packets type.
func (p *Puback) Type() Type {
	return PUBACK
}

// Len returns the byte length of the encoded packet.
func (p *Puback) Len(m Mode) int {
	return identifiedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Puback) Decode(m Mode, src []byte) (int, error) {
	return identifiedDecode(m, src, &p.ID, PUBACK)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Puback) Encode(m Mode, dst []byte) (int, error) {
	return identifiedEncode(m, dst, p.ID, PUBACK)
}

// String returns a string representation of the packet.
func (p *Puback) String() string {
	return fmt.Sprintf("<Puback ID=%d>", p.ID)
}

// A Pubcomp packet is the response to a Pubrel. It is the fourth and
// final packet of the QOS 2 protocol exchange.
type Pubcomp struct {
	// The packet identifier.
	ID ID

	ReasonCode byte

	Properties []Property
	// ReasonString string
	// UserProperties map[string][]byte
}

// NewPubcomp creates a new Pubcomp packet.
func NewPubcomp() *Pubcomp {
	return &Pubcomp{}
}

// Type returns the packets type.
func (p *Pubcomp) Type() Type {
	return PUBCOMP
}

// Len returns the byte length of the encoded packet.
func (p *Pubcomp) Len(m Mode) int {
	return identifiedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pubcomp) Decode(m Mode, src []byte) (int, error) {
	return identifiedDecode(m, src, &p.ID, PUBCOMP)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pubcomp) Encode(m Mode, dst []byte) (int, error) {
	return identifiedEncode(m, dst, p.ID, PUBCOMP)
}

// String returns a string representation of the packet.
func (p *Pubcomp) String() string {
	return fmt.Sprintf("<Pubcomp ID=%d>", p.ID)
}

// A Pubrec packet is the response to a Publish packet with QOS 2. It is the
// second packet of the QOS 2 protocol exchange.
type Pubrec struct {
	// Shared packet identifier.
	ID ID

	ReasonCode byte

	Properties []Property
	// ReasonString string
	// UserProperties map[string][]byte
}

// NewPubrec creates a new Pubrec packet.
func NewPubrec() *Pubrec {
	return &Pubrec{}
}

// Type returns the packets type.
func (p *Pubrec) Type() Type {
	return PUBREC
}

// Len returns the byte length of the encoded packet.
func (p *Pubrec) Len(m Mode) int {
	return identifiedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pubrec) Decode(m Mode, src []byte) (int, error) {
	return identifiedDecode(m, src, &p.ID, PUBREC)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pubrec) Encode(m Mode, dst []byte) (int, error) {
	return identifiedEncode(m, dst, p.ID, PUBREC)
}

// String returns a string representation of the packet.
func (p *Pubrec) String() string {
	return fmt.Sprintf("<Pubrec ID=%d>", p.ID)
}

// A Pubrel packet is the response to a Pubrec packet. It is the third packet of
// the QOS 2 protocol exchange.
type Pubrel struct {
	// Shared packet identifier.
	ID ID

	ReasonCode byte

	Properties []Property
	// ReasonString string
	// UserProperties map[string][]byte
}

// NewPubrel creates a new Pubrel packet.
func NewPubrel() *Pubrel {
	return &Pubrel{}
}

// Type returns the packets type.
func (p *Pubrel) Type() Type {
	return PUBREL
}

// Len returns the byte length of the encoded packet.
func (p *Pubrel) Len(m Mode) int {
	return identifiedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pubrel) Decode(m Mode, src []byte) (int, error) {
	return identifiedDecode(m, src, &p.ID, PUBREL)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pubrel) Encode(m Mode, dst []byte) (int, error) {
	return identifiedEncode(m, dst, p.ID, PUBREL)
}

// String returns a string representation of the packet.
func (p *Pubrel) String() string {
	return fmt.Sprintf("<Pubrel ID=%d>", p.ID)
}

// An Unsuback packet is sent by the server to the client to confirm receipt of
// an Unsubscribe packet.
type Unsuback struct {
	// Shared packet identifier.
	ID ID

	ReasonCode byte

	Properties []Property
	// ReasonString string
	// UserProperties map[string][]byte
}

// NewUnsuback creates a new Unsuback packet.
func NewUnsuback() *Unsuback {
	return &Unsuback{}
}

// Type returns the packets type.
func (u *Unsuback) Type() Type {
	return UNSUBACK
}

// Len returns the byte length of the encoded packet.
func (u *Unsuback) Len(m Mode) int {
	return identifiedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (u *Unsuback) Decode(m Mode, src []byte) (int, error) {
	return identifiedDecode(m, src, &u.ID, UNSUBACK)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (u *Unsuback) Encode(m Mode, dst []byte) (int, error) {
	return identifiedEncode(m, dst, u.ID, UNSUBACK)
}

// String returns a string representation of the packet.
func (u *Unsuback) String() string {
	return fmt.Sprintf("<Unsuback ID=%d>", u.ID)
}
