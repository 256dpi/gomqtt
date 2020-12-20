package packet

import (
	"fmt"
	"strings"
)

// A Suback packet is sent by the server to the client to confirm receipt and
// processing of a Subscribe packet. The Suback packet contains a list of return
// codes, that specify the maximum QOS levels that have been granted.
type Suback struct {
	// The granted QOS levels for the requested subscriptions.
	ReturnCodes []QOS

	// The packet identifier.
	ID ID
}

// NewSuback creates a new Suback packet.
func NewSuback() *Suback {
	return &Suback{}
}

// Type returns the packets type.
func (s *Suback) Type() Type {
	return SUBACK
}

// String returns a string representation of the packet.
func (s *Suback) String() string {
	var codes []string

	for _, c := range s.ReturnCodes {
		codes = append(codes, fmt.Sprintf("%d", c))
	}

	return fmt.Sprintf("<Suback ID=%d ReturnCodes=[%s]>",
		s.ID, strings.Join(codes, ", "))
}

// Len returns the byte length of the encoded packet.
func (s *Suback) Len() int {
	ml := s.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (s *Suback) Decode(src []byte) (int, error) {
	// decode header
	total, _, rl, err := decodeHeader(src, SUBACK)
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, insufficientBufferSize(SUBACK)
	}

	// check remaining length
	if rl <= 2 {
		return total, makeError(SUBACK, "expected remaining length to be greater than 2, got %d", rl)
	}

	// read packet id
	pid, n, err := readUint(src[total:], 2, SUBACK)
	total += n
	if err != nil {
		return total, err
	}

	// set packet id
	s.ID = ID(pid)
	if !s.ID.Valid() {
		return total, makeError(SUBACK, "packet id must be grater than zero")
	}

	// calculate number of return codes
	rcl := rl - 2

	// read return codes
	s.ReturnCodes = make([]QOS, rcl)
	for i, rc := range src[total : total+rcl] {
		s.ReturnCodes[i] = QOS(rc)
		total++
	}

	// validate return codes
	for i, code := range s.ReturnCodes {
		if !code.Successful() && code != QOSFailure {
			return total, makeError(SUBACK, "invalid return code %d for topic %d", code, i)
		}
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (s *Suback) Encode(dst []byte) (int, error) {
	// check return codes
	for i, code := range s.ReturnCodes {
		if !code.Successful() && code != QOSFailure {
			return 0, makeError(SUBACK, "invalid return code %d for topic %d", code, i)
		}
	}

	// check packet id
	if !s.ID.Valid() {
		return 0, makeError(SUBACK, "packet id must be grater than zero")
	}

	// encode header
	total, err := encodeHeader(dst, 0, s.len(), s.Len(), SUBACK)
	if err != nil {
		return total, err
	}

	// write packet id
	n, err := writeUint(dst[total:], uint64(s.ID), 2, SUBACK)
	total += n
	if err != nil {
		return total, err
	}

	// write return codes
	for _, rc := range s.ReturnCodes {
		dst[total] = byte(rc)
		total++
	}

	return total, nil
}

// Returns the payload length.
func (s *Suback) len() int {
	return 2 + len(s.ReturnCodes)
}
