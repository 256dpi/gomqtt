package packet

import "errors"

// ErrRemainingLengthMismatch is returned if the remaining length does not match
// the buffer length.
var ErrRemainingLengthMismatch = errors.New("remaining length mismatch")

// ErrInvalidStaticFlags is returned if the static flags do not match.
var ErrInvalidStaticFlags = errors.New("invalid static flags")

func headerLen(rl int) int {
	// add packet type and flag byte + remaining length
	return 1 + varintLen(uint64(rl))
}

func encodeHeader(dst []byte, flags byte, rl int, t Type) (int, error) {
	// check type
	if !t.Valid() {
		return 0, ErrInvalidPacketType
	}

	// check flags
	if flags != 0 && t != PUBLISH && flags != t.defaultFlags() {
		return 0, ErrInvalidStaticFlags
	}

	// write type and flags
	typeAndFlags := byte(t)<<4 | (t.defaultFlags() & 0b1111) | flags
	total, err := writeUint8(dst, typeAndFlags)
	if err != nil {
		return total, err
	}

	// write remaining length
	n, err := writeVarint(dst[1:], uint64(rl))
	total += n
	if err != nil {
		return total, err
	}

	// check buffer
	if rl != len(dst[total:]) {
		return total, ErrRemainingLengthMismatch
	}

	return total, nil
}

func decodeHeader(src []byte, t Type) (int, byte, int, error) {
	// read type and flags
	typeAndFlags, total, err := readUint8(src)
	if err != nil {
		return total, 0, 0, err
	}

	// read type and flags
	typ := Type(typeAndFlags >> 4)
	flags := typeAndFlags & 0b1111

	// check against static type
	if typ != t {
		return total, 0, 0, ErrInvalidPacketType
	}

	// check flags except for publish packets
	if t != PUBLISH && flags != t.defaultFlags() {
		return total, 0, 0, ErrInvalidStaticFlags
	}

	// read remaining length
	_rl, n, err := readVarint(src[total:])
	total += n
	if err != nil {
		return total, 0, 0, err
	}

	// get remaining length
	rl := int(_rl)

	// check buffer
	if rl != len(src[total:]) {
		return total, 0, 0, ErrRemainingLengthMismatch
	}

	return total, flags, rl, nil
}
