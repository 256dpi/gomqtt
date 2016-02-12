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
	"sync"

	"github.com/gomqtt/topic"
)

type Backend interface {
	Subscribe(*Client, string)
	Unsubscribe(*Client, string)
	Remove(*Client)
	Publish(*Client, string, []byte)

	StoreRetained(*Client, string, []byte)
	// TODO: support streaming of retained messages
	RetrieveRetained(*Client, string) []byte
}

// TODO: missing offline subscriptions

type MemoryBackend struct {
	tree     *topic.Tree
	retained map[string][]byte

	mutex sync.Mutex
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tree: topic.NewTree(),
	}
}

func (m *MemoryBackend) Subscribe(client *Client, filter string) {
	m.tree.Add(filter, client)
}

func (m *MemoryBackend) Unsubscribe(client *Client, filter string) {
	m.tree.Remove(filter, client)
}

func (m *MemoryBackend) Remove(client *Client) {
	m.tree.Clear(client)
}

func (m *MemoryBackend) Publish(client *Client, topic string, payload []byte) {
	for _, v := range m.tree.Match(topic) {
		v.(*Client).Publish(topic, payload)
	}
}

func (m *MemoryBackend) StoreRetained(client *Client, topic string, payload []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.retained[topic] = payload
}

func (m *MemoryBackend) RetrieveRetained(client *Client, topic string) []byte {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	payload, ok := m.retained[topic]

	if ok {
		return payload
	}

	return nil
}
