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

	MaxQOS            QOS
	NoLocal           bool
	RetainAsPublished bool
	RetainHandling    byte
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

	Properties []Property
	// SubscriptionIdentifier uint64
	// UserProperties map[string][]byte
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

	return fmt.Sprintf(
		"<Subscribe ID=%d Subscriptions=[%s]>",
		s.ID, strings.Join(subscriptions, ", "),
	)
}

// Len returns the byte length of the encoded packet.
func (s *Subscribe) Len(m Mode) int {
	ml := s.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (s *Subscribe) Decode(m Mode, src []byte) (int, error) {
	// decode header
	total, _, rl, err := decodeHeader(src, SUBSCRIBE)
	if err != nil {
		return total, wrapError(SUBSCRIBE, DECODE, m, total, err)
	}

	// read packet id
	pid, n, err := readUint(src[total:], 2)
	total += n
	if err != nil {
		return total, wrapError(SUBSCRIBE, DECODE, m, total, err)
	}

	// set packet id
	s.ID = ID(pid)
	if !s.ID.Valid() {
		return total, wrapError(SUBSCRIBE, DECODE, m, total, ErrInvalidPacketID)
	}

	// reset subscriptions
	s.Subscriptions = s.Subscriptions[:0]

	// calculate number of subscriptions
	sl := rl - 2

	// read subscriptions
	for sl > 0 {
		// read topic
		topic, n, err := readString(src[total:])
		total += n
		if err != nil {
			return total, wrapError(SUBSCRIBE, DECODE, m, total, err)
		}

		// check topic
		if len(topic) == 0 {
			return total, makeError(SUBSCRIBE, DECODE, m, total, "invalid topic")
		}

		// read qos
		_qos, n, err := readUint8(src[total:])
		total += n
		if err != nil {
			return total, wrapError(SUBSCRIBE, DECODE, m, total, err)
		}

		// get qos
		qos := QOS(_qos)
		if !qos.Successful() {
			return total, wrapError(SUBSCRIBE, DECODE, m, total, ErrInvalidQOSLevel)
		}

		// add subscription
		s.Subscriptions = append(s.Subscriptions, Subscription{Topic: topic, QOS: qos})

		// decrement counter
		sl -= 2 + len(topic) + 1
	}

	// check for empty subscription list
	if len(s.Subscriptions) == 0 {
		return total, makeError(SUBSCRIBE, DECODE, m, total, "missing subscriptions")
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (s *Subscribe) Encode(m Mode, dst []byte) (int, error) {
	// encode header
	total, err := encodeHeader(dst, 0, s.len(), SUBSCRIBE)
	if err != nil {
		return total, wrapError(SUBSCRIBE, ENCODE, m, total, err)
	}

	// check packet id
	if !s.ID.Valid() {
		return total, wrapError(SUBSCRIBE, ENCODE, m, total, ErrInvalidPacketID)
	}

	// write packet id
	n, err := writeUint(dst[total:], uint64(s.ID), 2)
	total += n
	if err != nil {
		return total, wrapError(SUBSCRIBE, ENCODE, m, total, err)
	}

	// check for empty subscription list
	if len(s.Subscriptions) == 0 {
		return total, makeError(SUBSCRIBE, ENCODE, m, total, "missing subscriptions")
	}

	// write subscriptions
	for _, sub := range s.Subscriptions {
		// check topic
		if len(sub.Topic) == 0 {
			return total, makeError(SUBSCRIBE, ENCODE, m, total, "invalid topic")
		}

		// write topic
		n, err = writeString(dst[total:], sub.Topic)
		total += n
		if err != nil {
			return total, wrapError(SUBSCRIBE, ENCODE, m, total, err)
		}

		// check qos
		if !sub.QOS.Successful() {
			return total, wrapError(SUBSCRIBE, ENCODE, m, total, ErrInvalidQOSLevel)
		}

		// write qos
		n, err = writeUint8(dst[total:], uint8(sub.QOS))
		total += n
		if err != nil {
			return total, wrapError(SUBSCRIBE, ENCODE, m, total, err)
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
