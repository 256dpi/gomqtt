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
	"bytes"
	"fmt"
	"sync/atomic"
)

// An UNSUBSCRIBE Packet is sent by the Client to the Server, to unsubscribe from topics.
type UnsubscribeMessage struct {
	header

	topics [][]byte
}

var _ Message = (*UnsubscribeMessage)(nil)

// NewUnsubscribeMessage creates a new UNSUBSCRIBE message.
func NewUnsubscribeMessage() *UnsubscribeMessage {
	msg := &UnsubscribeMessage{}
	msg.setType(UNSUBSCRIBE)

	return msg
}

func (this UnsubscribeMessage) String() string {
	msgstr := fmt.Sprintf("%s", this.header)

	for i, t := range this.topics {
		msgstr = fmt.Sprintf("%s, Topic%d=%s", msgstr, i, string(t))
	}

	return msgstr
}

// Topics returns a list of topics sent by the Client.
func (this *UnsubscribeMessage) Topics() [][]byte {
	return this.topics
}

// AddTopic adds a single topic to the message.
func (this *UnsubscribeMessage) AddTopic(topic []byte) {
	if this.TopicExists(topic) {
		return
	}

	this.topics = append(this.topics, topic)
}

// RemoveTopic removes a single topic from the list of existing ones in the message.
// If topic does not exist it just does nothing.
func (this *UnsubscribeMessage) RemoveTopic(topic []byte) {
	var i int
	var t []byte
	var found bool

	for i, t = range this.topics {
		if bytes.Equal(t, topic) {
			found = true
			break
		}
	}

	if found {
		this.topics = append(this.topics[:i], this.topics[i+1:]...)
	}
}

// TopicExists checks to see if a topic exists in the list.
func (this *UnsubscribeMessage) TopicExists(topic []byte) bool {
	for _, t := range this.topics {
		if bytes.Equal(t, topic) {
			return true
		}
	}

	return false
}

func (this *UnsubscribeMessage) Len() int {
	ml := this.msglen()

	if err := this.setRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return this.header.msglen() + ml
}

func (this *UnsubscribeMessage) Decode(src []byte) (int, error) {
	total := 0

	hn, err := this.header.decode(src[total:])
	total += hn
	if err != nil {
		return total, err
	}

	this.packetId = src[total : total+2]
	total += 2

	remlen := int(this.remlen) - (total - hn)
	for remlen > 0 {
		t, n, err := readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		this.topics = append(this.topics, t)
		remlen = remlen - n - 1
	}

	if len(this.topics) == 0 {
		return 0, fmt.Errorf(this.Name() + "/Decode: Empty topic list")
	}

	return total, nil
}

func (this *UnsubscribeMessage) Encode(dst []byte) (int, error) {
	hl := this.header.msglen()
	ml := this.msglen()

	if len(dst) < hl+ml {
		return 0, fmt.Errorf(this.Name() + "/Encode: Insufficient buffer size. Expecting %d, got %d.", hl+ml, len(dst))
	}

	if err := this.setRemainingLength(int32(ml)); err != nil {
		return 0, err
	}

	total := 0

	n, err := this.header.encode(dst[total:])
	total += n
	if err != nil {
		return total, err
	}

	if this.PacketId() == 0 {
		this.SetPacketId(uint16(atomic.AddUint64(&gPacketId, 1) & 0xffff))
	}

	n = copy(dst[total:], this.packetId)
	total += n

	for _, t := range this.topics {
		n, err := writeLPBytes(dst[total:], t)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (this *UnsubscribeMessage) msglen() int {
	// packet ID
	total := 2

	for _, t := range this.topics {
		total += 2 + len(t)
	}

	return total
}
