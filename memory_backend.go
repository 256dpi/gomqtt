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

	"github.com/gomqtt/session"
	"github.com/gomqtt/topic"
)

type MemoryBackend struct {
	tree     *topic.Tree
	retained map[string][]byte
	sessions map[string]*session.MemorySession

	mutex sync.Mutex
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tree:     topic.NewTree(),
		sessions: make(map[string]*session.MemorySession),
	}
}

func (m *MemoryBackend) GetSession(id string) (session.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sess, ok := m.sessions[id]
	if ok {
		return sess, nil
	}

	sess = session.NewMemorySession()
	m.sessions[id] = sess

	return sess, nil
}

func (m *MemoryBackend) Subscribe(client *Client, filter string) error {
	m.tree.Add(filter, client)

	return nil
}

func (m *MemoryBackend) Unsubscribe(client *Client, filter string) error {
	m.tree.Remove(filter, client)

	return nil
}

func (m *MemoryBackend) Remove(client *Client) error {
	m.tree.Clear(client)

	return nil
}

func (m *MemoryBackend) Publish(client *Client, topic string, payload []byte) error {
	for _, v := range m.tree.Match(topic) {
		// we do not care about errors here as it is not the publishing clients
		// responsibility
		v.(*Client).Publish(topic, payload)
	}

	return nil
}

func (m *MemoryBackend) StoreRetained(client *Client, topic string, payload []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.retained[topic] = payload

	return nil
}

func (m *MemoryBackend) RetrieveRetained(client *Client, topic string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	payload, ok := m.retained[topic]

	if ok {
		return payload, nil
	}

	return nil, nil
}
