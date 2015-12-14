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

package broker

import (
	"github.com/gomqtt/topic"
	"github.com/gomqtt/packet"
)

type QueueBackend interface {
	Subscribe(*Connection, string)
	Unsubscribe(*Connection, string)
	Publish(*Message)
}

type RetainedBackend interface {
	StoreRetained(*Connection, *Message)
	RetrieveRetained(*Connection, string) []*Message
}

type WillBackend interface {
	StoreWill(*Connection, *Message)
	RetrieveWill(*Connection) *Message
	ClearWill(*Connection)
}

// TODO: missing offline subscriptions

type MemoryBackend struct {
	tree *topic.Tree
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tree: topic.NewTree(),
	}
}

func (m *MemoryBackend) Subscribe(conn *Connection, filter string) {
	m.tree.Add(filter, conn)
}

func (m *MemoryBackend) Unsubscribe(conn *Connection, filter string) {
	m.tree.Remove(filter, conn)
}

func (m *MemoryBackend) Publish(message *Message) {
	for _, v := range m.tree.Match(message.Topic) {
		conn, ok := v.(*Connection)

		if ok {
			m := packet.NewPublishPacket()
			m.Topic = []byte(message.Topic)
			m.Payload = message.Payload
			m.QOS = message.QOS
			m.Retain = message.Retain
			conn.stream.Send(m)
		}
	}
}

