package packet

func headerLen(rl int) int {
	// add packet type and flag byte + remaining length
	return 1 + varintLen(uint64(rl))
}

func encodeHeader(dst []byte, flags byte, rl int, tl int, t Type) (int, error) {
	// check buffer length
	if len(dst) < headerLen(rl) || len(dst) < tl {
		return 0, insufficientBufferSize(t)
	}

	// write type and flags
	typeAndFlags := byte(t)<<4 | (t.defaultFlags() & 0xf)
	typeAndFlags |= flags
	dst[0] = typeAndFlags

	// write remaining length
	n, err := writeVarint(dst[1:], uint64(rl), t)
	if err != nil {
		return 0, err
	}

	return 1 + n, nil
}

func decodeHeader(src []byte, t Type) (int, byte, int, error) {
	// check buffer size
	if len(src) < 2 {
		return 0, 0, 0, insufficientBufferSize(t)
	}

	// read type and flags
	decodedType := Type(src[0] >> 4)
	flags := src[0] & 0x0f
	total := 1

	// check against static type
	if decodedType != t {
		return total, 0, 0, makeError(t, "invalid type %d", decodedType)
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

	// check remaining buffer
	if rl > len(src[total:]) {
		return total, 0, 0, makeError(t, "remaining length (%d) is greater than remaining buffer (%d)", rl, len(src[total:]))
	}

	return total, flags, rl, nil
}
