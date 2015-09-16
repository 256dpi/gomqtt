// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package message

import (
	"encoding/binary"
	"fmt"
)

// A single subscription in a SubscribeMessage.
type Subscription struct {
	// The topic to subscribe to.
	Topic []byte

	// The qos level for receiving the messages.
	QoS byte
}

// The SUBSCRIBE Packet is sent from the Client to the Server to create one or more
// Subscriptions. Each Subscription registers a Client’s interest in one or more
// Topics. The Server sends PUBLISH Packets to the Client in order to forward
// Application Messages that were published to Topics that match these Subscriptions.
// The SUBSCRIBE Packet also specifies (for each Subscription) the maximum QoS with
// which the Server can send Application Messages to the Client.
type SubscribeMessage struct {
	// The subscriptions.
	Subscriptions []Subscription

	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*SubscribeMessage)(nil)

// NewSubscribeMessage creates a new SUBSCRIBE message.
func NewSubscribeMessage() *SubscribeMessage {
	return &SubscribeMessage{}
}

// Type return the messages message type.
func (this SubscribeMessage) Type() MessageType {
	return SUBSCRIBE
}

// String returns a string representation of the message.
func (this SubscribeMessage) String() string {
	msgstr := fmt.Sprintf("SUBSCRIBE: PacketId=%d", this.PacketId)

	for i, t := range this.Subscriptions {
		msgstr = fmt.Sprintf("%s Topic[%d]=%q/%d", msgstr, i, string(t.Topic), t.QoS)
	}

	return msgstr
}

// Len returns the byte length of the message.
func (this *SubscribeMessage) Len() int {
	ml := this.msglen()
	return headerLen(ml) + ml
}

// Decode reads the bytes in the byte slice from the argument. It returns the
// total number of bytes decoded, and whether there have been any errors during
// the process. The byte slice MUST NOT be modified during the duration of this
// message being available since the byte slice never gets copied.
func (this *SubscribeMessage) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src[total:], SUBSCRIBE)
	total += hl
	if err != nil {
		return total, err
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("SUBSCRIBE/Decode: Insufficient buffer size. Expecting %d, got %d.", total+2, len(src))
	}

	// check remaining length
	if rl <= 2 {
		return total, fmt.Errorf("SUBSCRIBE/Decode: Expected remaining length to be greater that 2, got.", rl)
	}

	// read packet id
	this.PacketId = binary.BigEndian.Uint16(src[total:])
	total += 2

	// calculate number of subscriptions
	sl := int(rl) - 2

	for sl > 0 {
		// read topic
		t, n, err := readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		// check buffer length
		if len(src) < total+1 {
			return total, fmt.Errorf("SUBSCRIBE/Decode: Insufficient buffer size. Expecting %d, got %d.", total+1, len(src))
		}

		// read qos and add subscription
		this.Subscriptions = append(this.Subscriptions, Subscription{t, src[total]})
		total++

		// decrement counter
		sl = sl - n - 1
	}

	// check for empty subscription list
	if len(this.Subscriptions) == 0 {
		return total, fmt.Errorf("SUBSCRIBE/Decode: Empty subscription list.")
	}

	return total, nil
}

// Encode writes the message bytes into the byte array from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there's any errors, then the byte slice and count should be
// considered invalid.
func (this *SubscribeMessage) Encode(dst []byte) (int, error) {
	total := 0

	// check buffer length
	l := this.Len()
	if len(dst) < l {
		return total, fmt.Errorf("SUBSCRIBE/Encode: Insufficient buffer size. Expecting %d, got %d.", l, len(dst))
	}

	// encode header
	n, err := headerEncode(dst[total:], 0, this.msglen(), SUBSCRIBE)
	total += n
	if err != nil {
		return total, err
	}

	// write packet it
	binary.BigEndian.PutUint16(dst[total:], this.PacketId)
	total += 2

	for _, t := range this.Subscriptions {
		// write topic
		n, err := writeLPBytes(dst[total:], t.Topic)
		total += n
		if err != nil {
			return total, err
		}

		// write qos
		dst[total] = t.QoS

		total++
	}

	return total, nil
}

// Returns the payload length.
func (this *SubscribeMessage) msglen() int {
	// packet ID
	total := 2

	for _, t := range this.Subscriptions {
		total += 2 + len(t.Topic) + 1
	}

	return total
}
