package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

// ErrInsufficientBufferSize is returned if the provided buffer is too small
// to write to or read from.
var ErrInsufficientBufferSize = errors.New("insufficient buffer size")

// ErrVariableIntegerOverflow is returned if the provided or to be written
// integer requires more than 4 bytes to write or read.
var ErrVariableIntegerOverflow = errors.New("variable integer overflow")

// ErrPrefixedBytesOverflow is returned if the provided bytes require more than
// two bytes to encode its length.
var ErrPrefixedBytesOverflow = errors.New("prefixed bytes overflow")

// ErrInvalidUTF8Sequence is returned if a string is not a valid utf8 sequence.
var ErrInvalidUTF8Sequence = errors.New("invalid utf8 sequence")

func readUint8(buf []byte) (uint8, int, error) {
	num, n, err := readUint(buf, 1)
	return uint8(num), n, err
}

func writeUint8(buf []byte, num uint8) (int, error) {
	return writeUint(buf, uint64(num), 1)
}

func readUint(buf []byte, width int) (uint64, int, error) {
	// check buffer
	if len(buf) < width {
		return 0, 0, ErrInsufficientBufferSize
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
		panic("unsupported width")
	}

	return num, width, nil
}

func writeUint(buf []byte, num uint64, width int) (int, error) {
	// check buffer
	if len(buf) < width {
		return 0, ErrInsufficientBufferSize
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
		panic("unsupported width")
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

func readVarint(buf []byte) (uint64, int, error) {
	// read number
	num, n := binary.Uvarint(buf)
	if n == 0 {
		return 0, 0, ErrInsufficientBufferSize
	} else if n < 0 || n > 4 {
		return 0, 0, ErrVariableIntegerOverflow
	}

	return num, n, nil
}

func writeVarint(buf []byte, num uint64) (int, error) {
	// check size
	if num > maxVarint {
		return 0, ErrVariableIntegerOverflow
	}

	// check length
	if len(buf) < varintLen(num) {
		return 0, ErrInsufficientBufferSize
	}

	// write remaining length
	n := binary.PutUvarint(buf, num)

	return n, nil
}

func readString(buf []byte) (string, int, error) {
	// read str
	str, n, err := readBytes(buf, false)
	if err != nil {
		return "", n, err
	}

	// check validity
	if !utf8.Valid(str) || bytes.ContainsRune(str, 0) {
		return "", n, ErrInvalidUTF8Sequence
	}

	return string(str), n, nil
}

func writeString(buf []byte, str string) (int, error) {
	// check string
	if !utf8.ValidString(str) || strings.ContainsRune(str, 0) {
		return 0, ErrInvalidUTF8Sequence
	}

	return writeBytes(buf, cast(str))
}

func readBytes(buf []byte, safe bool) ([]byte, int, error) {
	// read length
	l, n, err := readUint(buf, 2)
	if err != nil {
		return nil, n, err
	}

	// get length
	length := int(l)

	// check length
	if len(buf[n:]) < length {
		return nil, n, ErrInsufficientBufferSize
	}

	// get bytes
	bytes := buf[n : n+length]

	// return input buffer if not safe
	if !safe {
		return bytes, n + length, nil
	}

	// otherwise copy buffer
	cpy := make([]byte, length)
	n += copy(cpy, bytes)

	return cpy, n, nil
}

func writeBytes(buf []byte, bytes []byte) (int, error) {
	// get length
	length := len(bytes)

	// check length
	if length > math.MaxUint16 {
		return 0, ErrPrefixedBytesOverflow
	}

	// write length
	n, err := writeUint(buf, uint64(length), 2)
	if err != nil {
		return n, err
	}

	// check buffer
	if len(buf) < length {
		return n, ErrInsufficientBufferSize
	}

	// write bytes
	n += copy(buf[2:], bytes)

	return n, nil
}

func readPair(buf []byte) (string, string, int, error) {
	// read key
	key, nk, err := readString(buf)
	if err != nil {
		return "", "", nk, err
	}

	// read value
	value, nv, err := readString(buf[nk:])
	if err != nil {
		return key, "", nk + nv, err
	}

	return key, value, nk + nv, nil
}

func writePair(buf []byte, key, value string) (int, error) {
	// write key
	nk, err := writeString(buf, key)
	if err != nil {
		return nk, err
	}

	// write value
	nv, err := writeString(buf, value)
	if err != nil {
		return nk + nv, err
	}

	return nk + nv, nil
}
