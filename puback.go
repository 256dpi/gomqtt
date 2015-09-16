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

// A PUBACK Packet is the response to a PUBLISH Packet with QoS level 1.
type PubackMessage struct {
	// Shared message identifier.
	PacketId uint16
}

var _ Message = (*PubackMessage)(nil)

// NewPubackMessage creates a new PUBACK message.
func NewPubackMessage() *PubackMessage {
	msg := &PubackMessage{}
	return msg
}

func (this PubackMessage) Type() MessageType {
	return PUBACK
}

func (this *PubackMessage) Len() int {
	return identifiedMessageLen()
}

func (this *PubackMessage) Decode(src []byte) (int, error) {
	n, pid, err := identifiedMessageDecode(src, PUBACK)
	this.PacketId = pid
	return n, err
}

func (this *PubackMessage) Encode(dst []byte) (int, error) {
	return identifiedMessageEncode(dst, this.PacketId, PUBACK)
}
