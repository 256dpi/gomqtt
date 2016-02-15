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

	"github.com/gomqtt/packet"
	"github.com/gomqtt/session"
	"github.com/gomqtt/topic"
)

type MemoryBackend struct {
	queue    *topic.Tree
	retained *topic.Tree
	sessions map[string]*session.MemorySession

	mutex sync.Mutex
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		queue:    topic.NewTree(),
		retained: topic.NewTree(),
		sessions: make(map[string]*session.MemorySession),
	}
}

func (m *MemoryBackend) GetSession(client *Client, id string) (session.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(id) > 0 {
		sess, ok := m.sessions[id]
		if ok {
			return sess, nil
		}

		sess = session.NewMemorySession()
		m.sessions[id] = sess

		return sess, nil
	}

	return session.NewMemorySession(), nil
}

func (m *MemoryBackend) Subscribe(client *Client, topic string) ([]*packet.Message, error) {
	m.queue.Add(topic, client)

	values := m.retained.Search(topic)

	var msgs []*packet.Message

	for _, value := range values {
		if msg, ok := value.(*packet.Message); ok {
			msgs = append(msgs, msg)
		}
	}

	return msgs, nil
}

func (m *MemoryBackend) Unsubscribe(client *Client, topic string) error {
	m.queue.Remove(topic, client)

	return nil
}

func (m *MemoryBackend) Remove(client *Client) error {
	m.queue.Clear(client)

	return nil
}

func (m *MemoryBackend) Publish(client *Client, msg *packet.Message) error {
	if msg.Retain {
		if len(msg.Payload) > 0 {
			m.retained.Set(msg.Topic, msg)
		} else {
			m.retained.Empty(msg.Topic)
		}
	}

	for _, v := range m.queue.Match(msg.Topic) {
		if client, ok := v.(*Client); ok && client != nil {
			client.Publish(msg)
		}
	}

	return nil
}
