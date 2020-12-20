package packet

import "fmt"

// A Publish packet is sent from a client to a server or from server to a client
// to transport an application message.
type Publish struct {
	// The message to publish.
	Message Message

	// If the Dup flag is set to false, it indicates that this is the first
	// occasion that the client or server has attempted to send this
	// Publish packet. If the dup flag is set to true, it indicates that this
	// might be re-delivery of an earlier attempt to send the packet.
	Dup bool

	// The packet identifier.
	ID ID
}

// NewPublish creates a new Publish packet.
func NewPublish() *Publish {
	return &Publish{}
}

// Type returns the packets type.
func (p *Publish) Type() Type {
	return PUBLISH
}

// String returns a string representation of the packet.
func (p *Publish) String() string {
	return fmt.Sprintf("<Publish ID=%d Message=%s Dup=%t>",
		p.ID, p.Message.String(), p.Dup)
}

// Len returns the byte length of the encoded packet.
func (p *Publish) Len() int {
	ml := p.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of
// bytes decoded, and whether there have been any errors during the process.
func (p *Publish) Decode(src []byte) (int, error) {
	// decode header
	hl, flags, rl, err := decodeHeader(src, PUBLISH)
	total := hl
	if err != nil {
		return total, err
	}

	// read flags
	p.Dup = ((flags >> 3) & 0x1) == 1
	p.Message.Retain = (flags & 0x1) == 1
	p.Message.QOS = QOS((flags >> 1) & 0x3)

	// check qos
	if !p.Message.QOS.Successful() {
		return total, makeError(PUBLISH, "invalid QOS level (%d)", p.Message.QOS)
	}

	// read topic
	topic, n, err := readLPString(src[total:], PUBLISH)
	total += n
	if err != nil {
		return total, err
	}

	// set topic
	p.Message.Topic = topic

	// check quality of service
	if p.Message.QOS != 0 {
		// check buffer length
		if len(src) < total+2 {
			return total, insufficientBufferSize(PUBLISH)
		}

		// read packet id
		pid, n, err := readUint(src[total:], 2, PUBLISH)
		total += n
		if err != nil {
			return total, err
		}

		// set packet id
		p.ID = ID(pid)
		if !p.ID.Valid() {
			return total, makeError(PUBLISH, "packet id must be grater than zero")
		}
	}

	// calculate payload length
	l := rl - (total - hl)

	// read payload
	if l > 0 {
		p.Message.Payload = make([]byte, l)
		copy(p.Message.Payload, src[total:total+l])
		total += len(p.Message.Payload)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (p *Publish) Encode(dst []byte) (int, error) {
	// check topic length
	if len(p.Message.Topic) == 0 {
		return 0, makeError(PUBLISH, "topic name is empty")
	}

	// prepare flags
	var flags byte

	// set dup flag
	if p.Dup {
		flags |= 0x8 // 00001000
	}

	// set retain flag
	if p.Message.Retain {
		flags |= 0x1 // 00000001
	}

	// check qos
	if !p.Message.QOS.Successful() {
		return 0, makeError(PUBLISH, "invalid QOS level %d", p.Message.QOS)
	}

	// check packet id
	if p.Message.QOS > 0 && !p.ID.Valid() {
		return 0, makeError(PUBLISH, "packet id must be grater than zero")
	}

	// set qos
	flags = (flags & 249) | (byte(p.Message.QOS) << 1) // 249 = 11111001

	// encode header
	total, err := encodeHeader(dst, flags, p.len(), p.Len(), PUBLISH)
	if err != nil {
		return total, err
	}

	// write topic
	n, err := writeLPString(dst[total:], p.Message.Topic, PUBLISH)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	if p.Message.QOS != 0 {
		n, err := writeUint(dst[total:], uint64(p.ID), 2, PUBLISH)
		total += n
		if err != nil {
			return total, err
		}
	}

	// write payload
	copy(dst[total:], p.Message.Payload)
	total += len(p.Message.Payload)

	return total, nil
}

func (p *Publish) len() int {
	// topic + payload
	total := 2 + len(p.Message.Topic) + len(p.Message.Payload)

	// packet iD
	if p.Message.QOS != 0 {
		total += 2
	}

	return total
}
