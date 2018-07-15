package packet

import (
	"encoding/binary"
	"fmt"
)

// Returns the byte length of an identified packet.
func identifiedPacketLen() int {
	return headerLen(2) + 2
}

// Decodes an identified packet.
func identifiedPacketDecode(src []byte, t Type) (int, ID, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, t)
	total += hl
	if err != nil {
		return total, 0, err
	}

	// check remaining length
	if rl != 2 {
		return total, 0, fmt.Errorf("[%s] expected remaining length to be 2", t)
	}

	// read packet id
	packetID := binary.BigEndian.Uint16(src[total:])
	total += 2

	// check packet id
	if packetID == 0 {
		return total, 0, fmt.Errorf("[%s] packet id must be grater than zero", t)
	}

	return total, ID(packetID), nil
}

// Encodes an identified packet.
func identifiedPacketEncode(dst []byte, id ID, t Type) (int, error) {
	total := 0

	// check packet id
	if id == 0 {
		return total, fmt.Errorf("[%s] packet id must be grater than zero", t)
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, identifiedPacketLen(), t)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	binary.BigEndian.PutUint16(dst[total:], uint16(id))
	total += 2

	return total, nil
}

// A PubackPacket is the response to a PublishPacket with QOS level 1.
type PubackPacket struct {
	// The packet identifier.
	ID ID
}

// NewPubackPacket creates a new PubackPacket.
func NewPubackPacket() *PubackPacket {
	return &PubackPacket{}
}

// Type returns the packets type.
func (pp *PubackPacket) Type() Type {
	return PUBACK
}

// Len returns the byte length of the encoded packet.
func (pp *PubackPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubackPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBACK)
	pp.ID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubackPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.ID, PUBACK)
}

// String returns a string representation of the packet.
func (pp *PubackPacket) String() string {
	return fmt.Sprintf("<PubackPacket ID=%d>", pp.ID)
}

// A PubcompPacket is the response to a PubrelPacket. It is the fourth and
// final packet of the QOS 2 protocol exchange.
type PubcompPacket struct {
	// The packet identifier.
	ID ID
}

var _ Generic = (*PubcompPacket)(nil)

// NewPubcompPacket creates a new PubcompPacket.
func NewPubcompPacket() *PubcompPacket {
	return &PubcompPacket{}
}

// Type returns the packets type.
func (pp *PubcompPacket) Type() Type {
	return PUBCOMP
}

// Len returns the byte length of the encoded packet.
func (pp *PubcompPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubcompPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBCOMP)
	pp.ID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubcompPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.ID, PUBCOMP)
}

// String returns a string representation of the packet.
func (pp *PubcompPacket) String() string {
	return fmt.Sprintf("<PubcompPacket ID=%d>", pp.ID)
}

// A PubrecPacket is the response to a PublishPacket with QOS 2. It is the
// second packet of the QOS 2 protocol exchange.
type PubrecPacket struct {
	// Shared packet identifier.
	ID ID
}

// NewPubrecPacket creates a new PubrecPacket.
func NewPubrecPacket() *PubrecPacket {
	return &PubrecPacket{}
}

// Type returns the packets type.
func (pp *PubrecPacket) Type() Type {
	return PUBREC
}

// Len returns the byte length of the encoded packet.
func (pp *PubrecPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubrecPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBREC)
	pp.ID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubrecPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.ID, PUBREC)
}

// String returns a string representation of the packet.
func (pp *PubrecPacket) String() string {
	return fmt.Sprintf("<PubrecPacket ID=%d>", pp.ID)
}

// A PubrelPacket is the response to a PubrecPacket. It is the third packet of
// the QOS 2 protocol exchange.
type PubrelPacket struct {
	// Shared packet identifier.
	ID ID
}

var _ Generic = (*PubrelPacket)(nil)

// NewPubrelPacket creates a new PubrelPacket.
func NewPubrelPacket() *PubrelPacket {
	return &PubrelPacket{}
}

// Type returns the packets type.
func (pp *PubrelPacket) Type() Type {
	return PUBREL
}

// Len returns the byte length of the encoded packet.
func (pp *PubrelPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (pp *PubrelPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, PUBREL)
	pp.ID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pp *PubrelPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, pp.ID, PUBREL)
}

// String returns a string representation of the packet.
func (pp *PubrelPacket) String() string {
	return fmt.Sprintf("<PubrelPacket ID=%d>", pp.ID)
}

// An UnsubackPacket is sent by the server to the client to confirm receipt of
// an UnsubscribePacket.
type UnsubackPacket struct {
	// Shared packet identifier.
	ID ID
}

// NewUnsubackPacket creates a new UnsubackPacket.
func NewUnsubackPacket() *UnsubackPacket {
	return &UnsubackPacket{}
}

// Type returns the packets type.
func (up *UnsubackPacket) Type() Type {
	return UNSUBACK
}

// Len returns the byte length of the encoded packet.
func (up *UnsubackPacket) Len() int {
	return identifiedPacketLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (up *UnsubackPacket) Decode(src []byte) (int, error) {
	n, pid, err := identifiedPacketDecode(src, UNSUBACK)
	up.ID = pid
	return n, err
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (up *UnsubackPacket) Encode(dst []byte) (int, error) {
	return identifiedPacketEncode(dst, up.ID, UNSUBACK)
}

// String returns a string representation of the packet.
func (up *UnsubackPacket) String() string {
	return fmt.Sprintf("<UnsubackPacket ID=%d>", up.ID)
}
