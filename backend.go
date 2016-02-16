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

// A Consumer is a client connected to the broker that interacts with the backend.
type Consumer interface {
	// Publish will send a Message to the consumer and initiate QOS flows.
	Publish(msg *packet.Message) bool

	// Context returns the associated context.
	Context() *Context
}

// A Backend provides effective queuing functionality to a Broker and its Clients.
type Backend interface {
	// Authenticate authenticates a consumers credentials.
	Authenticate(consumer Consumer, user, password string) (bool, error)

	// GetSession returns the already stored session for the supplied id or creates
	// and returns a new one. If the supplied id has a zero length, a new
	// session is returned that is only valid once. The Backend may also allocate
	// other resources and setup the consumer.
	GetSession(consumer Consumer, id string) (Session, error)

	// Subscribe will subscribe the passed consumer to the specified topic and
	// begin to forward messages by calling the consumers Publish method.
	// It will also return the stored retained messages matching the supplied
	// topic.
	Subscribe(consumer Consumer, topic string) ([]*packet.Message, error)

	// Unsubscribe will unsubscribe the passed consumer from the specified topic.
	Unsubscribe(consumer Consumer, topic string) error

	// Publish will forward the passed message to all other subscribed consumers.
	// It will also store the message if Retain is set to true. If the supplied
	//message has additionally a zero length payload, the backend removes the
	// currently retained message.
	Publish(consumer Consumer, msg *packet.Message) error

	// Remove will unsubscribe the passed consumer from previously subscribe topics.
	// The Backend may also cleanup previously allocated resources for that consumer.
	Remove(consumer Consumer) error
}

// TODO: missing offline subscriptions

// AbstractBackendAuthenticationTest tests a backend implementations Authenticate
// method. The backend should allow the "allow:allow" and deny the "deny:deny"
// logins.
func AbstractBackendAuthenticationTest(t *testing.T, backend Backend) {
	ok, err := backend.Authenticate(nil, "allow", "allow")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = backend.Authenticate(nil, "deny", "deny")
	assert.False(t, ok)
	assert.NoError(t, err)
}

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

// A MemoryBackend stores everything in memory.
type MemoryBackend struct {
	queue    *tools.Tree
	retained *tools.Tree

	Logins map[string]string

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

// Authenticate authenticates a consumers credentials.
func (m *MemoryBackend) Authenticate(consumer Consumer, user, password string) (bool, error) {
	// allow all if there are no logins
	if m.Logins == nil {
		return true, nil
	}

	// check login
	if pw, ok := m.Logins[user]; ok && pw == password {
		return true, nil
	}

	return false, nil
}

// GetSession returns the already stored session for the supplied id or creates
// and returns a new one. If the supplied id has a zero length, a new
// session is returned that is only valid once.
func (m *MemoryBackend) GetSession(consumer Consumer, id string) (Session, error) {
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

// Subscribe will subscribe the passed consumer to the specified topic and
// begin to forward messages by calling the consumers Publish method.
// It will also return the stored retained messages matching the supplied
// topic.
func (m *MemoryBackend) Subscribe(consumer Consumer, topic string) ([]*packet.Message, error) {
	m.queue.Add(topic, consumer)

	values := m.retained.Search(topic)

	var msgs []*packet.Message

	for _, value := range values {
		if msg, ok := value.(*packet.Message); ok {
			msgs = append(msgs, msg)
		}
	}

	return msgs, nil
}

// Unsubscribe will unsubscribe the passed consumer from the specified topic.
func (m *MemoryBackend) Unsubscribe(consumer Consumer, topic string) error {
	m.queue.Remove(topic, consumer)
	return nil
}

// Publish will forward the passed message to all other subscribed consumers.
// It will also store the message if Retain is set to true. If the supplied
//message has additionally a zero length payload, the backend removes the
// currently retained message.
func (m *MemoryBackend) Publish(consumer Consumer, msg *packet.Message) error {
	if msg.Retain {
		if len(msg.Payload) > 0 {
			m.retained.Set(msg.Topic, msg)
		} else {
			m.retained.Empty(msg.Topic)
		}
	}

	for _, v := range m.queue.Match(msg.Topic) {
		if consumer, ok := v.(Consumer); ok && consumer != nil {
			consumer.Publish(msg)
		}
	}

	return nil
}

// Remove will unsubscribe the passed consumer from previously subscribe topics.
func (m *MemoryBackend) Remove(consumer Consumer) error {
	m.queue.Clear(consumer)
	return nil
}
