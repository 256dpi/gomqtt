package packet

import (
	"fmt"
	"strings"
)

// A Subscription is a single subscription in a Subscribe packet.
type Subscription struct {
	// The topic to subscribe.
	Topic string

	// The requested maximum QOS level.
	QOS QOS
}

func (s *Subscription) String() string {
	return fmt.Sprintf("%q=>%d", s.Topic, s.QOS)
}

// A Subscribe packet is sent from the client to the server to create one or
// more Subscriptions. The server will forward application messages that match
// these subscriptions using PublishPackets.
type Subscribe struct {
	// The subscriptions.
	Subscriptions []Subscription

	// The packet identifier.
	ID ID
}

// NewSubscribe creates a new Subscribe packet.
func NewSubscribe() *Subscribe {
	return &Subscribe{}
}

// Type returns the packets type.
func (s *Subscribe) Type() Type {
	return SUBSCRIBE
}

// String returns a string representation of the packet.
func (s *Subscribe) String() string {
	// collect subscriptions
	var subscriptions []string
	for _, t := range s.Subscriptions {
		subscriptions = append(subscriptions, t.String())
	}

	return fmt.Sprintf("<Subscribe ID=%d Subscriptions=[%s]>",
		s.ID, strings.Join(subscriptions, ", "))
}

// Len returns the byte length of the encoded packet.
func (s *Subscribe) Len() int {
	ml := s.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (s *Subscribe) Decode(src []byte) (int, error) {
	// decode header
	total, _, rl, err := decodeHeader(src, SUBSCRIBE)
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, insufficientBufferSize(SUBSCRIBE)
	}

	// read packet id
	pid, n, err := readUint(src[total:], 2, SUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// set packet id
	s.ID = ID(pid)
	if !s.ID.Valid() {
		return total, makeError(SUBSCRIBE, "packet id must be grater than zero")
	}

	// reset subscriptions
	s.Subscriptions = s.Subscriptions[:0]

	// calculate number of subscriptions
	sl := rl - 2

	// read subscriptions
	for sl > 0 {
		// read topic
		topic, n, err := readLPString(src[total:], SUBSCRIBE)
		total += n
		if err != nil {
			return total, err
		}

		// check buffer length
		if len(src) < total+1 {
			return total, makeError(SUBSCRIBE, "insufficient buffer size, expected %d, got %d", total+1, len(src))
		}

		// read qos
		_qos, n, err := readUint(src[total:], 1, SUBSCRIBE)
		total += n
		if err != nil {
			return total, err
		}

		// get qos
		qos := QOS(_qos)
		if !qos.Successful() {
			return total, makeError(SUBSCRIBE, "invalid QOS level (%d)", qos)
		}

		// add subscription
		s.Subscriptions = append(s.Subscriptions, Subscription{Topic: topic, QOS: qos})

		// decrement counter
		sl -= 2 + len(topic) + 1
	}

	// check for empty subscription list
	if len(s.Subscriptions) == 0 {
		return total, makeError(SUBSCRIBE, "empty subscription list")
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (s *Subscribe) Encode(dst []byte) (int, error) {
	// check packet id
	if !s.ID.Valid() {
		return 0, makeError(SUBSCRIBE, "packet id must be grater than zero")
	}

	// encode header
	total, err := encodeHeader(dst, 0, s.len(), s.Len(), SUBSCRIBE)
	if err != nil {
		return total, err
	}

	// write packet id
	n, err := writeUint(dst[total:], uint64(s.ID), 2, SUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// write subscriptions
	for _, sub := range s.Subscriptions {
		// write topic
		n, err = writeLPString(dst[total:], sub.Topic, SUBSCRIBE)
		total += n
		if err != nil {
			return total, err
		}

		// check qos
		if !sub.QOS.Successful() {
			return total, makeError(SUBSCRIBE, "invalid QOS level (%d)", sub.QOS)
		}

		// write qos
		n, err = writeUint(dst[total:], uint64(sub.QOS), 1, SUBSCRIBE)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (s *Subscribe) len() int {
	// packet ID
	total := 2

	// add subscriptions
	for _, t := range s.Subscriptions {
		total += 2 + len(t.Topic) + 1
	}

	return total
}
