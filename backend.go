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
	Remove(conn *Connection)
	Publish(*packet.PublishPacket)
}

type RetainedBackend interface {
	StoreRetained(*Connection, *packet.PublishPacket)
	RetrieveRetained(*Connection, string) []*packet.PublishPacket
}

type WillBackend interface {
	StoreWill(*Connection, *packet.PublishPacket)
	RetrieveWill(*Connection) *packet.PublishPacket
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

func (m *MemoryBackend) Remove(conn *Connection) {
	m.tree.Clear(conn)
}

func (m *MemoryBackend) Publish(message *packet.PublishPacket) {
	for _, v := range m.tree.Match(string(message.Topic)) {
		conn, ok := v.(*Connection)

		if ok {
			conn.stream.Send(message)
		}
	}
}

