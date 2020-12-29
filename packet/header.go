package packet

func headerLen(rl int) int {
	// add packet type and flag byte + remaining length
	return 1 + varintLen(uint64(rl))
}

func encodeHeader(dst []byte, flags byte, rl int, t Type) (int, error) {
	// write type and flags
	typeAndFlags := byte(t)<<4 | (t.defaultFlags() & 0xf) | flags
	total, err := writeUint8(dst, typeAndFlags, t)
	if err != nil {
		return total, err
	}

	// write remaining length
	n, err := writeVarint(dst[1:], uint64(rl), t)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func decodeHeader(src []byte, t Type) (int, byte, int, error) {
	// read type and flags
	typeAndFlags, total, err := readUint8(src, t)
	if err != nil {
		return total, 0, 0, err
	}

	// read type and flags
	typ := Type(typeAndFlags >> 4)
	flags := typeAndFlags & 0x0f

	// check against static type
	if typ != t {
		return total, 0, 0, makeError(t, "invalid type %d", typ)
	}

	// check flags except for publish packets
	if t != PUBLISH && flags != t.defaultFlags() {
		return total, 0, 0, makeError(t, "invalid flags, expected %d, got %d", t.defaultFlags(), flags)
	}

	// read remaining length
	_rl, n, err := readVarint(src[total:], t)
	total += n
	if err != nil {
		return total, 0, 0, err
	}

	// get remaining length
	rl := int(_rl)

	// check buffer
	if rl != len(src[total:]) {
		return total, 0, 0, makeError(t, "remaining length (%d) does not equal remaining buffer size (%d)", rl, len(src[total:]))
	}

	return total, flags, rl, nil
}
