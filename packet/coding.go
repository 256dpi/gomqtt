package packet

import (
	"encoding/binary"
	"math"
)

const maxVarint = 268435455

func varintLen(n uint64) int {
	if n < 128 {
		return 1
	} else if n < 16384 {
		return 2
	} else if n < 2097152 {
		return 3
	} else if n <= maxVarint {
		return 4
	}

	return 0
}

func readVarint(buf []byte, t Type) (uint64, int, error) {
	// limit buf
	if len(buf) > 4 {
		buf = buf[:4]
	}

	// read number
	num, n := binary.Uvarint(buf)
	if n <= 0 {
		return 0, 0, makeError(t, "error reading variable integer")
	}

	// check size
	if num > maxVarint {
		return 0, n, makeError(t, "variable integer (%d) out of bound (max %d, min 0)", num, maxVarint)
	}

	return num, n, nil
}

func writeVarint(buf []byte, num uint64, t Type) (int, error) {
	// check size
	if num > maxVarint {
		return 0, makeError(t, "variable integer out of bound")
	}

	// check length
	if len(buf) < varintLen(num) {
		return 0, makeError(t, "insufficient buffer size")
	}

	// write remaining length
	n := binary.PutUvarint(buf, num)

	return n, nil
}

func readLPBytes(buf []byte, safe bool, t Type) ([]byte, int, error) {
	// check buffer
	if len(buf) < 2 {
		return nil, 0, makeError(t, "insufficient buffer size, expected 2, got %d", len(buf))
	}

	// read length
	length := int(binary.BigEndian.Uint16(buf))

	// check length
	if len(buf) < 2+length {
		return nil, 2, makeError(t, "insufficient buffer size, expected %d, got %d", 2+length, len(buf))
	}

	// get bytes
	bytes := buf[2 : 2+length]

	// return input buffer if not safe
	if !safe {
		return bytes, 2 + length, nil
	}

	// otherwise copy buffer
	cpy := make([]byte, length)
	copy(cpy, bytes)

	return cpy, 2 + length, nil
}

func readLPString(buf []byte, t Type) (string, int, error) {
	bytes, n, err := readLPBytes(buf, false, t)
	return string(bytes), n, err
}

func writeLPBytes(buf []byte, bytes []byte, t Type) (int, error) {
	// get length
	length := len(bytes)

	// check length
	if length > math.MaxUint16 {
		return 0, makeError(t, "length %d greater than allowed %d bytes", length, math.MaxUint16)
	}

	// check buffer
	if len(buf) < 2+length {
		return 0, makeError(t, "insufficient buffer size, expected %d, got %d", 2+length, len(buf))
	}

	// write length
	binary.BigEndian.PutUint16(buf, uint16(length))

	// write bytes
	copy(buf[2:], bytes)

	return 2 + length, nil
}

func writeLPString(buf []byte, str string, t Type) (int, error) {
	return writeLPBytes(buf, []byte(str), t)
}
