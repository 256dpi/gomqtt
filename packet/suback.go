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

	ReasonCode byte

	Properties []Property
	// ReasonString string
	// UserProperties map[string][]byte
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
	// collect return codes
	var returnCodes []string
	for _, c := range s.ReturnCodes {
		returnCodes = append(returnCodes, fmt.Sprintf("%d", c))
	}

	return fmt.Sprintf(
		"<Suback ID=%d ReturnCodes=[%s]>",
		s.ID, strings.Join(returnCodes, ", "),
	)
}

// Len returns the byte length of the encoded packet.
func (s *Suback) Len(m Mode) int {
	ml := s.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (s *Suback) Decode(m Mode, src []byte) (int, error) {
	// decode header
	total, _, rl, err := decodeHeader(src, SUBACK)
	if err != nil {
		return total, wrapError(SUBACK, DECODE, m, err)
	}

	// read packet id
	pid, n, err := readUint(src[total:], 2)
	total += n
	if err != nil {
		return total, wrapError(SUBACK, DECODE, m, err)
	}

	// set packet id
	s.ID = ID(pid)
	if !s.ID.Valid() {
		return total, wrapError(SUBACK, DECODE, m, ErrInvalidPacketID)
	}

	// calculate number of return codes
	rcl := rl - 2
	if rcl < 1 {
		return total, makeError(SUBACK, DECODE, m, "expected at least one return code")
	}

	// prepare return codes
	s.ReturnCodes = make([]QOS, 0, rcl)

	// read return codes
	for i := 0; i < rcl; i++ {
		// read return code
		rc, n, err := readUint8(src[total:])
		total += n
		if err != nil {
			return total, wrapError(SUBACK, DECODE, m, err)
		}

		// get return code
		returnCode := QOS(rc)
		if !returnCode.Successful() && returnCode != QOSFailure {
			return total, makeError(SUBACK, DECODE, m, "invalid return code %d", returnCode)
		}

		// add return code
		s.ReturnCodes = append(s.ReturnCodes, returnCode)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (s *Suback) Encode(m Mode, dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, s.len(), SUBACK)
	if err != nil {
		return total, wrapError(SUBACK, ENCODE, m, err)
	}

	// check packet id
	if !s.ID.Valid() {
		return total, wrapError(SUBACK, ENCODE, m, ErrInvalidPacketID)
	}

	// write packet id
	n, err := writeUint(dst[total:], uint64(s.ID), 2)
	total += n
	if err != nil {
		return total, wrapError(SUBACK, ENCODE, m, err)
	}

	// write return codes
	for _, rc := range s.ReturnCodes {
		// check return code
		if !rc.Successful() && rc != QOSFailure {
			return total, makeError(SUBACK, ENCODE, m, "invalid return code %d", rc)
		}

		// write return code
		n, err := writeUint8(dst[total:], uint8(rc))
		total += n
		if err != nil {
			return total, wrapError(SUBACK, ENCODE, m, err)
		}
	}

	return total, nil
}

func (s *Suback) len() int {
	return 2 + len(s.ReturnCodes)
}
