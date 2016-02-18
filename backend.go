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

// A Backend provides effective queuing functionality to a Broker and its Consumers.
type Backend interface {
	// Authenticate should authenticate the consumer using the user and password
	// values and return true if the consumer is eligible to continue or false
	// when the broker should terminate the connection.
	Authenticate(consumer Consumer, user, password string) (bool, error)

	// Setup is called when a new consumer comes online and is successfully
	// authenticated. Setup should return the already stored session for the
	// supplied id or create and return a new one. If clean is set to true it
	// should additionally reset the session. If the supplied id has a zero
	// length, a new session is returned that is not stored further.
	//
	// Note: In this call the Backend may also allocate other resources and
	// setup the consumer for further usage as the broker will acknowledge the
	// connection when the call returns.
	Setup(consumer Consumer, id string, clean bool) (Session, bool, error)

	// Subscribe should subscribe the passed consumer to the specified topic and
	// call Publish with any incoming messages. It should also return the stored
	// retained messages that match the specified topic. Additionally, the Backend
	// may start another goroutine in the background that publishes missed QOS 1
	// and QOS 2 messages to the consumer.
	Subscribe(consumer Consumer, topic string) ([]*packet.Message, error)

	// Unsubscribe should unsubscribe the passed consumer from the specified topic.
	Unsubscribe(consumer Consumer, topic string) error

	// Publish should forward the passed message to all other consumers that hold
	// a subscription that matches the messages topic. It should also store the
	// message if Retain is set to true and the payload does not have a zero
	// length. If the payload has a zero length and Retain is set to true the
	// currently retained message for that topic should be removed.
	Publish(consumer Consumer, msg *packet.Message) error

	// Terminate is called when the consumer goes offline. Terminate should
	// unsubscribe the passed consumer from all previously subscribed topics.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that consumer as the broker will close the connection when the call
	// returns.
	Terminate(consumer Consumer) error
}

// AbstractBackendAuthenticationTest tests a backend implementations Authenticate
// method. The backend should allow the "allow:allow" and deny the "deny:deny"
// logins.
func AbstractBackendAuthenticationTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	ok, err := backend.Authenticate(consumer, "allow", "allow")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = backend.Authenticate(consumer, "deny", "deny")
	assert.False(t, ok)
	assert.NoError(t, err)
}

// AbstractBackendSetupTest tests a backend implementations Setup method.
func AbstractBackendSetupTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	// has id and clean=false

	session1, resumed, err := backend.Setup(consumer, "foo", false)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session1)

	session2, resumed, err := backend.Setup(consumer, "foo", false)
	assert.NoError(t, err)
	assert.True(t, resumed)
	assert.True(t, session1 == session2)

	// has id and clean=true

	session3, resumed, err := backend.Setup(consumer, "foo", true)
	assert.NoError(t, err)
	assert.True(t, resumed)
	assert.True(t, session1 == session3)

	// has other id and clean=false

	session4, resumed, err := backend.Setup(consumer, "bar", false)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.True(t, session4 != session1)
	assert.True(t, session4 != session2)
	assert.True(t, session4 != session3)

	// has no id and clean=true

	session5, resumed, err := backend.Setup(consumer, "", true)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session5)

	session6, resumed, err := backend.Setup(consumer, "", true)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session5)
	assert.True(t, session5 != session6)
}

// AbstractQueuingTest tests a backend implementations queuing methods.
func AbstractQueuingTest(t *testing.T, backend Backend) {
	consumer1 := newFakeConsumer()
	consumer2 := newFakeConsumer()

	msg := &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
	}

	// subscribe both consumers

	msgs, err := backend.Subscribe(consumer1, "test")
	assert.Nil(t, msgs)
	assert.NoError(t, err)

	msgs, err = backend.Subscribe(consumer2, "test")
	assert.Nil(t, msgs)
	assert.NoError(t, err)

	// publish message

	err = backend.Publish(consumer1, msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, consumer1.in[0])
	assert.Equal(t, msg, consumer2.in[0])

	// unsubscribe one consumer

	err = backend.Unsubscribe(consumer2, "test")
	assert.NoError(t, err)

	// publish another message

	err = backend.Publish(consumer1, msg)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(consumer1.in))
	assert.Equal(t, 1, len(consumer2.in))
}

// AbstractBackendRetainedTest tests a backend implementations message retaining.
func AbstractBackendRetainedTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

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
	msgs, err := backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)

	err = backend.Publish(consumer, msg1)
	assert.NoError(t, err)

	// should have one
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg1, msgs[0])

	err = backend.Publish(consumer, msg2)
	assert.NoError(t, err)

	// should have two
	msgs, err = backend.Subscribe(consumer, "#")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(msgs))

	err = backend.Publish(consumer, msg3)
	assert.NoError(t, err)

	// should have another
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg3, msgs[0])

	err = backend.Publish(consumer, msg4)
	assert.NoError(t, err)

	// should have none
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

// A MemoryBackend stores everything in memory.
type MemoryBackend struct {
	Logins map[string]string

	queue         *tools.Tree
	retained      *tools.Tree
	offlineQueue  *tools.Tree
	sessions      map[string]*MemorySession
	sessionsMutex sync.Mutex
}

// NewMemoryBackend returns a new MemoryBackend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		queue:        tools.NewTree(),
		retained:     tools.NewTree(),
		offlineQueue: tools.NewTree(),
		sessions:     make(map[string]*MemorySession),
	}
}

// Authenticate authenticates a consumers credentials by matching them to the
// saved Logins map.
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

// Setup returns the already stored session for the supplied id or creates
// and returns a new one. If clean is set to true it will additionally reset
// the session. If the supplied id has a zero length, a new session is returned
// that is not stored further.
func (m *MemoryBackend) Setup(consumer Consumer, id string, clean bool) (Session, bool, error) {
	m.sessionsMutex.Lock()
	defer m.sessionsMutex.Unlock()

	// save clean flag
	consumer.Context().Set("clean", clean)

	// check id length
	if len(id) > 0 {
		sess, ok := m.sessions[id]

		// when found
		if ok {
			// reset session if clean is true
			if clean {
				sess.Reset()
			}

			// remove all session from the offline queue
			m.offlineQueue.Clear(sess)

			// send all missed messages in another goroutine
			go func(){
				for _, msg := range sess.missed() {
					consumer.Publish(msg)
				}
			}()

			// returned stored session
			consumer.Context().Set("session", sess)
			return sess, true, nil
		}

		// create fresh session and save it
		sess = NewMemorySession()
		m.sessions[id] = sess

		// return new stored session
		consumer.Context().Set("session", sess)
		return sess, false, nil
	}

	// return a new temporary session
	sess := NewMemorySession()
	consumer.Context().Set("session", sess)
	return sess, false, nil
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
// message has additionally a zero length payload, the backend removes the
// currently retained message.
func (m *MemoryBackend) Publish(consumer Consumer, msg *packet.Message) error {
	if msg.Retain {
		if len(msg.Payload) > 0 {
			m.retained.Set(msg.Topic, msg)
		} else {
			m.retained.Empty(msg.Topic)
		}
	}

	// publish directly to consumers
	for _, v := range m.queue.Match(msg.Topic) {
		if consumer, ok := v.(Consumer); ok {
			consumer.Publish(msg)
		}
	}

	// queue for offline consumers
	for _, v := range m.offlineQueue.Match(msg.Topic) {
		if session, ok := v.(*MemorySession); ok {
			session.queue(msg)
		}
	}

	return nil
}

// Terminate will unsubscribe the passed consumer from all previously subscribed topics.
func (m *MemoryBackend) Terminate(consumer Consumer) error {
	m.queue.Clear(consumer)

	// get session
	session, ok := consumer.Context().Get("session").(*MemorySession)
	if ok {
		// check if the consumer connected with clean=true
		clean, ok := consumer.Context().Get("clean").(bool)
		if ok && clean {
			// reset session
			session.Reset()
			return nil
		}

		// otherwise get stored subscriptions
		subscriptions, err := session.AllSubscriptions()
		if err != nil {
			return err
		}

		// iterate through stored subscriptions
		for _, sub := range subscriptions {
			if sub.QOS >= 1 {
				// session to offline queue
				m.offlineQueue.Add(sub.Topic, session)
			}
		}
	}

	return nil
}
