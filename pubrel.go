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
import "fmt"

// A PUBREL Packet is the response to a PUBREC Packet. It is the third packet of the
// QoS 2 protocol exchange.
type PubrelMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubrelMessage)(nil)

// NewPubrelMessage creates a new PUBREL message.
func NewPubrelMessage() *PubrelMessage {
	msg := &PubrelMessage{}
	return msg
}

func (this PubrelMessage) Type() MessageType {
	return PUBREL
}


func (this *PubrelMessage) Len() int {
	return identifiedMessageLen()
}

func (this *PubrelMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBREL)
	this.PacketId = pid
	return n, err
}

func (this *PubrelMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, this.PacketId, PUBREL)
}

// String returns a string representation of the message.
func (this PubrelMessage) String() string {
	return fmt.Sprintf("PUBREL: PacketId=%d", this.PacketId)
}
