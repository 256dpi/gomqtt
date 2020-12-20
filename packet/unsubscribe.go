package packet

import (
	"fmt"
	"strings"
)

// An Unsubscribe packet is sent by the client to the server.
type Unsubscribe struct {
	// The topics to unsubscribe from.
	Topics []string

	// The packet identifier.
	ID ID
}

// NewUnsubscribe creates a new Unsubscribe packet.
func NewUnsubscribe() *Unsubscribe {
	return &Unsubscribe{}
}

// Type returns the packets type.
func (u *Unsubscribe) Type() Type {
	return UNSUBSCRIBE
}

// String returns a string representation of the packet.
func (u *Unsubscribe) String() string {
	// collect topics
	var topics []string
	for _, t := range u.Topics {
		topics = append(topics, fmt.Sprintf("%q", t))
	}

	return fmt.Sprintf("<Unsubscribe Topics=[%s]>", strings.Join(topics, ", "))
}

// Len returns the byte length of the encoded packet.
func (u *Unsubscribe) Len() int {
	ml := u.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (u *Unsubscribe) Decode(src []byte) (int, error) {
	// decode header
	total, _, rl, err := decodeHeader(src, UNSUBSCRIBE)
	if err != nil {
		return total, err
	}

	// read packet id
	pid, n, err := readUint(src[total:], 2, UNSUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// set packet id
	u.ID = ID(pid)
	if !u.ID.Valid() {
		return total, makeError(UNSUBSCRIBE, "packet id must be grater than zero")
	}

	// reset topics
	u.Topics = u.Topics[:0]

	// read topics
	tl := rl - 2
	for tl > 0 {
		// read topic
		topic, n, err := readLPString(src[total:], UNSUBSCRIBE)
		total += n
		if err != nil {
			return total, err
		}

		// append to list
		u.Topics = append(u.Topics, topic)

		// decrement counter
		tl -= n
	}

	// check for empty list
	if len(u.Topics) == 0 {
		return total, makeError(UNSUBSCRIBE, "empty topic list")
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (u *Unsubscribe) Encode(dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, u.len(), u.Len(), UNSUBSCRIBE)
	if err != nil {
		return total, err
	}

	// check packet id
	if !u.ID.Valid() {
		return 0, makeError(UNSUBSCRIBE, "packet id must be grater than zero")
	}

	// write packet id
	n, err := writeUint(dst[total:], uint64(u.ID), 2, UNSUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// write topics
	for _, topic := range u.Topics {
		// write topic
		n, err := writeLPString(dst[total:], topic, UNSUBSCRIBE)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (u *Unsubscribe) len() int {
	// packet ID
	total := 2

	// add topics
	for _, t := range u.Topics {
		total += 2 + len(t)
	}

	return total
}
