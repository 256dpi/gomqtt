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
	"testing"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/stretchr/testify/assert"
)

// A Backend provides effective queueing functionality to a Broker and its Clients.
type Backend interface {
	// GetSession returns the already stored session for the supplied id or creates
	// and returns a new one. If the supplied id has a zero length, a new
	// session is returned that is only valid once.
	GetSession(client *Client, id string) (Session, error)

	// Subscribe will subscribe the passed client to the specified topic and
	// begin to forward messages by calling the clients Publish method.
	// It will also return the stored retained messages matching the supplied
	// topic.
	Subscribe(client *Client, topic string) ([]*packet.Message, error)

	// Unsubscribe will unsubscribe the passed client from the specified topic.
	Unsubscribe(client *Client, topic string) error

	// Remove will unsubscribe the passed client from previously subscribe topics.
	Remove(client *Client) error

	// Publish will forward the passed message to all other subscribed clients.
	// It will also store the message if Retain is set to true. If the supplied
	//message has additionally a zero length payload, the backend removes the
	// currently retained message.
	Publish(client *Client, msg *packet.Message) error
}

// TODO: missing offline subscriptions

// AbstractBackendGetSessionTest tests a backend implementations GetSession method.
func AbstractBackendGetSessionTest(t *testing.T, backend Backend) {
	session1, err := backend.GetSession(nil, "foo")
	assert.NoError(t, err)
	assert.NotNil(t, session1)

	session2, err := backend.GetSession(nil, "foo")
	assert.NoError(t, err)
	assert.True(t, session1 == session2)

	session3, err := backend.GetSession(nil, "bar")
	assert.NoError(t, err)
	assert.False(t, session3 == session1)
	assert.False(t, session3 == session2)

	session4, err := backend.GetSession(nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, session4)

	session5, err := backend.GetSession(nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, session5)
	assert.True(t, session4 != session5)
}

// AbstractBackendRetainedTest tests a backend implementations message retaining.
func AbstractBackendRetainedTest(t *testing.T, backend Backend) {
	msg1 := &packet.Message{
		Topic:   "foo",
		Payload: []byte("bar"),
		QOS:     1,
		Retain:  true,
	}

	msg2 := &packet.Message{
		Topic:   "foo/bar",
		Payload: []byte("bar"),
		QOS:     1,
		Retain:  true,
	}

	msg3 := &packet.Message{
		Topic:   "foo",
		Payload: []byte("bar"),
		QOS:     2,
		Retain:  true,
	}

	msg4 := &packet.Message{
		Topic:  "foo",
		QOS:    1,
		Retain: true,
	}

	// should be empty
	msgs, err := backend.Subscribe(nil, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)

	err = backend.Publish(nil, msg1)
	assert.NoError(t, err)

	// should have one
	msgs, err = backend.Subscribe(nil, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg1, msgs[0])

	err = backend.Publish(nil, msg2)
	assert.NoError(t, err)

	// should have two
	msgs, err = backend.Subscribe(nil, "#")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(msgs))

	err = backend.Publish(nil, msg3)
	assert.NoError(t, err)

	// should have another
	msgs, err = backend.Subscribe(nil, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg3, msgs[0])

	err = backend.Publish(nil, msg4)
	assert.NoError(t, err)

	// should have none
	msgs, err = backend.Subscribe(nil, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

// MemoryBackend stores everything in memory.
type MemoryBackend struct {
	queue    *tools.Tree
	retained *tools.Tree

	sessions      map[string]*MemorySession
	sessionsMutex sync.Mutex
}

// NewMemoryBackend returns a new MemoryBackend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		queue:    tools.NewTree(),
		retained: tools.NewTree(),
		sessions: make(map[string]*MemorySession),
	}
}

// GetSession returns the already stored session for the supplied id or creates
// and returns a new one. If the supplied id has a zero length, a new
// session is returned that is only valid once.
func (m *MemoryBackend) GetSession(client *Client, id string) (Session, error) {
	m.sessionsMutex.Lock()
	defer m.sessionsMutex.Unlock()

	if len(id) > 0 {
		sess, ok := m.sessions[id]
		if ok {
			return sess, nil
		}

		sess = NewMemorySession()
		m.sessions[id] = sess

		return sess, nil
	}

	return NewMemorySession(), nil
}

// Subscribe will subscribe the passed client to the specified topic and
// begin to forward messages by calling the clients Publish method.
// It will also return the stored retained messages matching the supplied
// topic.
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

// Unsubscribe will unsubscribe the passed client from the specified topic.
func (m *MemoryBackend) Unsubscribe(client *Client, topic string) error {
	m.queue.Remove(topic, client)
	return nil
}

// Remove will unsubscribe the passed client from previously subscribe topics.
func (m *MemoryBackend) Remove(client *Client) error {
	m.queue.Clear(client)
	return nil
}

// Publish will forward the passed message to all other subscribed clients.
// It will also store the message if Retain is set to true. If the supplied
//message has additionally a zero length payload, the backend removes the
// currently retained message.
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
