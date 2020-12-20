package packet

func nakedLen() int {
	return headerLen(0)
}

func nakedDecode(src []byte, t Type) (int, error) {
	// decode header
	hl, _, rl, err := decodeHeader(src, t)
	if rl != 0 {
		return hl, makeError(t, "expected zero remaining length")
	}

	return hl, err
}

func nakedEncode(dst []byte, t Type) (int, error) {
	return encodeHeader(dst, 0, 0, nakedLen(), t)
}

// A Disconnect packet is sent from the client to the server.
// It indicates that the client is disconnecting cleanly.
type Disconnect struct{}

// NewDisconnect creates a new Disconnect packet.
func NewDisconnect() *Disconnect {
	return &Disconnect{}
}

// Type returns the packets type.
func (d *Disconnect) Type() Type {
	return DISCONNECT
}

// Len returns the byte length of the encoded packet.
func (d *Disconnect) Len() int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (d *Disconnect) Decode(src []byte) (int, error) {
	return nakedDecode(src, DISCONNECT)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (d *Disconnect) Encode(dst []byte) (int, error) {
	return nakedEncode(dst, DISCONNECT)
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
func (p *Pingreq) Len() int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pingreq) Decode(src []byte) (int, error) {
	return nakedDecode(src, PINGREQ)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pingreq) Encode(dst []byte) (int, error) {
	return nakedEncode(dst, PINGREQ)
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
func (p *Pingresp) Len() int {
	return nakedLen()
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Pingresp) Decode(src []byte) (int, error) {
	return nakedDecode(src, PINGRESP)
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Pingresp) Encode(dst []byte) (int, error) {
	return nakedEncode(dst, PINGRESP)
}

// String returns a string representation of the packet.
func (p *Pingresp) String() string {
	return "<Pingresp>"
}
