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

// The PUBCOMP Packet is the response to a PUBREL Packet. It is the fourth and
// final packet of the QoS 2 protocol exchange.
type PubcompMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubcompMessage)(nil)

// NewPubcompMessage creates a new PUBCOMP message.
func NewPubcompMessage() *PubcompMessage {
	msg := &PubcompMessage{}
	return msg
}

func (this PubcompMessage) Type() MessageType {
	return PUBCOMP
}


func (this *PubcompMessage) Len() int {
	return identifiedMessageLen()
}

func (this *PubcompMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBCOMP)
	this.PacketId = pid
	return n, err
}

func (this *PubcompMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, this.PacketId, PUBCOMP)
}
