package packet

import (
	"encoding/binary"
	"math"
)

func readUint8(buf []byte, t Type) (uint8, int, error) {
	num, n, err := readUint(buf, 1, t)
	return uint8(num), n, err
}

func writeUint8(buf []byte, num uint8, t Type) (int, error) {
	return writeUint(buf, uint64(num), 1, t)
}

func readUint(buf []byte, width int, t Type) (uint64, int, error) {
	// check buffer
	if len(buf) < width {
		return 0, 0, insufficientBufferSize(t)
	}

	// read number
	var num uint64
	switch width {
	case 1:
		num = uint64(buf[0])
	case 2:
		num = uint64(binary.BigEndian.Uint16(buf))
	case 4:
		num = uint64(binary.BigEndian.Uint32(buf))
	case 8:
		num = binary.BigEndian.Uint64(buf)
	default:
		panic("unsupported uint width")
	}

	return num, width, nil
}

func writeUint(buf []byte, num uint64, width int, t Type) (int, error) {
	// check buffer
	if len(buf) < width {
		return 0, insufficientBufferSize(t)
	}

	// write number
	switch width {
	case 1:
		buf[0] = uint8(num)
	case 2:
		binary.BigEndian.PutUint16(buf, uint16(num))
	case 4:
		binary.BigEndian.PutUint32(buf, uint32(num))
	case 8:
		binary.BigEndian.PutUint64(buf, num)
	default:
		panic("unsupported uint width")
	}

	return width, nil
}

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

	return num, n, nil
}

func writeVarint(buf []byte, num uint64, t Type) (int, error) {
	// check size
	if num > maxVarint {
		return 0, makeError(t, "variable integer out of bound")
	}

	// check length
	if len(buf) < varintLen(num) {
		return 0, insufficientBufferSize(t)
	}

	// write remaining length
	n := binary.PutUvarint(buf, num)

	return n, nil
}

func readLPString(buf []byte, t Type) (string, int, error) {
	bytes, n, err := readLPBytes(buf, false, t)
	return string(bytes), n, err
}

func writeLPString(buf []byte, str string, t Type) (int, error) {
	return writeLPBytes(buf, cast(str), t)
}

func readLPBytes(buf []byte, safe bool, t Type) ([]byte, int, error) {
	// read length
	l, n, err := readUint(buf, 2, t)
	if err != nil {
		return nil, n, err
	}

	// get length
	length := int(l)

	// check length
	if len(buf[n:]) < length {
		return nil, n, insufficientBufferSize(t)
	}

	// get bytes
	bytes := buf[n : n+length]

	// return input buffer if not safe
	if !safe {
		return bytes, n + length, nil
	}

	// otherwise, copy buffer
	cpy := make([]byte, length)
	n += copy(cpy, bytes)

	return cpy, n, nil
}

func writeLPBytes(buf []byte, bytes []byte, t Type) (int, error) {
	// get length
	length := len(bytes)

	// check length
	if length > math.MaxUint16 {
		return 0, makeError(t, "length %d greater than allowed %d bytes", length, math.MaxUint16)
	}

	// write length
	n, err := writeUint(buf, uint64(length), 2, t)
	if err != nil {
		return n, err
	}

	// check buffer
	if len(buf) < length {
		return n, insufficientBufferSize(t)
	}

	// write bytes
	n += copy(buf[2:], bytes)

	return n, nil
}
