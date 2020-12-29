package packet

func nakedLen() int {
	return headerLen(0)
}

func nakedDecode(m Mode, src []byte, t Type) (int, error) {
	// decode header
	total, _, _, err := decodeHeader(src, t)
	if err != nil {
		return total, err
	}

	// check buffer
	if len(src[total:]) != 0 {
		return total, makeError(t, "leftover buffer length (%d)", len(src[total:]))
	}

	return total, nil
}

func nakedEncode(m Mode, dst []byte, t Type) (int, error) {
	return encodeHeader(dst, 0, 0, t)
}

type Auth struct {
	ReasonCode byte

	Properties []Property
	// AuthenticationMethod string
	// AuthenticationData []byte
	// ReasonString string
	// UserProperties map[string][]byte
}

// NewAuth creates a new Auth packet.
func NewAuth() *Auth {
	return &Auth{}
}

// Type returns the packets type.
func (ap *Auth) Type() Type {
	return AUTH
}

// Len returns the byte length of the encoded packet.
func (ap *Auth) Len(m Mode) int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (ap *Auth) Decode(m Mode, src []byte) (int, error) {
	return nakedDecode(m, src, AUTH)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (ap *Auth) Encode(m Mode, dst []byte) (int, error) {
	return nakedEncode(m, dst, AUTH)
}

// String returns a string representation of the packet.
func (ap *Auth) String() string {
	return "<Auth>"
}

// A Disconnect packet is sent from the client to the server.
// It indicates that the client is disconnecting cleanly.
type Disconnect struct {
	ReasonCode byte

	Properties []Property
	// SessionExpiryInterval uint64
	// ReasonString string
	// ServerReference string
	// UserProperties map[string][]byte
}

// NewDisconnect creates a new Disconnect packet.
func NewDisconnect() *Disconnect {
	return &Disconnect{}
}

// Type returns the packets type.
func (d *Disconnect) Type() Type {
	return DISCONNECT
}

// Len returns the byte length of the encoded packet.
func (d *Disconnect) Len(m Mode) int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (d *Disconnect) Decode(m Mode, src []byte) (int, error) {
	return nakedDecode(m, src, DISCONNECT)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (d *Disconnect) Encode(m Mode, dst []byte) (int, error) {
	return nakedEncode(m, dst, DISCONNECT)
}

// String returns a string representation of the packet.
func (d *Disconnect) String() string {
	return "<Disconnect>"
}

// A Pingreq packet is sent from a client to the server.
type Pingreq struct{}

// NewPingreq creates a new Pingreq packet.
func NewPingreq() *Pingreq {
	return &Pingreq{}
}

// Type returns the packets type.
func (p *Pingreq) Type() Type {
	return PINGREQ
}

// Len returns the byte length of the encoded packet.
func (p *Pingreq) Len(m Mode) int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pingreq) Decode(m Mode, src []byte) (int, error) {
	return nakedDecode(m, src, PINGREQ)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pingreq) Encode(m Mode, dst []byte) (int, error) {
	return nakedEncode(m, dst, PINGREQ)
}

// String returns a string representation of the packet.
func (p *Pingreq) String() string {
	return "<Pingreq>"
}

// A Pingresp packet is sent by the server to the client in response to a
// Pingreq. It indicates that the server is alive.
type Pingresp struct{}

// NewPingresp creates a new Pingresp packet.
func NewPingresp() *Pingresp {
	return &Pingresp{}
}

// Type returns the packets type.
func (p *Pingresp) Type() Type {
	return PINGRESP
}

// Len returns the byte length of the encoded packet.
func (p *Pingresp) Len(m Mode) int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pingresp) Decode(m Mode, src []byte) (int, error) {
	return nakedDecode(m, src, PINGRESP)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pingresp) Encode(m Mode, dst []byte) (int, error) {
	return nakedEncode(m, dst, PINGRESP)
}

// String returns a string representation of the packet.
func (p *Pingresp) String() string {
	return "<Pingresp>"
}
