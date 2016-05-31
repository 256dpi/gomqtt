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
	"github.com/gomqtt/tools"
)

// A Backend provides effective queuing functionality to a Broker and its Clients.
type Backend interface {
	// Authenticate should authenticate the client using the user and password
	// values and return true if the client is eligible to continue or false
	// when the broker should terminate the connection.
	Authenticate(client Client, user, password string) (bool, error)

	// Setup is called when a new client comes online and is successfully
	// authenticated. Setup should return the already stored session for the
	// supplied id or create and return a new one. If clean is set to true it
	// should additionally reset the session. If the supplied id has a zero
	// length, a new temporary session should returned that is not stored further.
	//
	// Optional: The backend may close any existing clients that use the same
	// client id.
	//
	// Note: In this call the Backend may also allocate other resources and
	// setup the client for further usage as the broker will acknowledge the
	// connection when the call returns.
	Setup(client Client, id string, clean bool) (Session, bool, error)

	// Restore is called after the client has been acknowledged and if clean is
	// set to false. It should be used to restore a clients subscriptions from
	// the session and triggering a background process that forwards all missed
	// messages.
	Restore(client Client) error

	// Subscribe should subscribe the passed client to the specified topic and
	// call Publish with any incoming messages.
	Subscribe(client Client, topic string) error

	// Unsubscribe should unsubscribe the passed client from the specified topic.
	Unsubscribe(client Client, topic string) error

	// StoreRetained should store the specified message.
	StoreRetained(client Client, msg *packet.Message) error

	// ClearRetained should remove the stored messages for the given topic.
	ClearRetained(client Client, topic string) error

	// QueueRetained is called after acknowledging a subscription and should be
	// used to trigger a background process that forwards all retained messages.
	QueueRetained(client Client, topic string) error

	// Publish should forward the passed message to all other clients that hold
	// a subscription that matches the messages topic. It should also add the
	// message to all sessions that have a matching offline subscription.
	Publish(client Client, msg *packet.Message) error

	// Terminate is called when the client goes offline. Terminate should
	// unsubscribe the passed client from all previously subscribed topics.
	//
	// Optional: The backend may convert a clients subscriptions into offline
	// subscriptions, which allows missed messages to be forwarded on the next
	// connect.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that client as the broker will close the connection when the call
	// returns.
	Terminate(client Client) error
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

// Authenticate authenticates a clients credentials by matching them to the
// saved Logins map.
func (m *MemoryBackend) Authenticate(client Client, user, password string) (bool, error) {
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
// that is not stored further. Furthermore, it will disconnect any client
// connected with the same client id.
func (m *MemoryBackend) Setup(client Client, id string, clean bool) (Session, bool, error) {
	m.sessionsMutex.Lock()
	defer m.sessionsMutex.Unlock()

	// save clean flag
	client.Context().Set("clean", clean)

	// return a new temporary session if id is zero
	if len(id) == 0 {
		sess := NewMemorySession()
		client.Context().Set("session", sess)
		return sess, false, nil
	}

	// retrieve stored session
	sess, ok := m.sessions[id]

	// when found
	if ok {
		// check if session already has a client
		if sess.currentClient != nil {
			sess.currentClient.Close(true)
		}

		// set current client
		sess.currentClient = client

		// reset session if clean is true
		if clean {
			sess.Reset()
		}

		// clear offline subscriptions
		m.offlineQueue.Clear(sess)

		// returned stored session
		client.Context().Set("session", sess)
		return sess, true, nil
	}

	// create fresh session
	sess = NewMemorySession()
	sess.currentClient = client

	// save session
	m.sessions[id] = sess

	// return new stored session
	client.Context().Set("session", sess)
	return sess, false, nil
}

// Restore will resubscribe the client to the saved subscriptions and forward
// all missed messages in another goroutine.
func (m *MemoryBackend) Restore(client Client) error {
	// get client session
	sess := client.Context().Get("session").(*MemorySession)

	// get stored subscriptions
	subs, err := sess.AllSubscriptions()
	if err != nil {
		return err
	}

	// add subscriptions
	for _, sub := range subs {
		m.queue.Add(sub.Topic, client)
	}

	// send all missed messages in another goroutine
	go func() {
		for _, msg := range sess.missed() {
			client.Publish(msg)
		}
	}()

	return nil
}

// Subscribe will subscribe the passed client to the specified topic and
// begin to forward messages by calling the clients Publish method.
func (m *MemoryBackend) Subscribe(client Client, topic string) error {
	// add subscription
	m.queue.Add(topic, client)

	return nil
}

// Unsubscribe will unsubscribe the passed client from the specified topic.
func (m *MemoryBackend) Unsubscribe(client Client, topic string) error {
	// remove subscription
	m.queue.Remove(topic, client)

	return nil
}

// StoreRetained will store the specified message.
func (m *MemoryBackend) StoreRetained(client Client, msg *packet.Message) error {
	// set retained message
	m.retained.Set(msg.Topic, msg.Copy())

	return nil
}

// ClearRetained will remove the stored messages for the given topic.
func (m *MemoryBackend) ClearRetained(client Client, topic string) error {
	// clear retained message
	m.retained.Empty(topic)

	return nil
}

// QueueRetained will queue all retained messages matching the given topic.
func (m *MemoryBackend) QueueRetained(client Client, topic string) error {
	// get retained messages
	values := m.retained.Search(topic)

	// publish messages
	for _, value := range values {
		client.Publish(value.(*packet.Message))
	}

	return nil
}

// Publish will forward the passed message to all other subscribed clients. It
// will also add the message to all sessions that have a matching offline
// subscription.
func (m *MemoryBackend) Publish(client Client, msg *packet.Message) error {
	// publish directly to clients
	for _, v := range m.queue.Match(msg.Topic) {
		v.(Client).Publish(msg)
	}

	// queue for offline clients
	for _, v := range m.offlineQueue.Match(msg.Topic) {
		v.(*MemorySession).queue(msg)
	}

	return nil
}

// Terminate will unsubscribe the passed client from all previously subscribed
// topics. If the client connect with clean=true it will also clean the session.
// Otherwise it will create offline subscriptions for all QOS 1 and QOS 2
// subscriptions.
func (m *MemoryBackend) Terminate(client Client) error {
	m.sessionsMutex.Lock()
	defer m.sessionsMutex.Unlock()

	// clear all subscriptions
	m.queue.Clear(client)

	// get session
	session, ok := client.Context().Get("session").(*MemorySession)
	if ok {
		// reset stored client
		session.currentClient = nil

		// check if the client has connected with clean=true
		clean, ok := client.Context().Get("clean").(bool)
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
				// add offline subscription
				m.offlineQueue.Add(sub.Topic, session)
			}
		}
	}

	return nil
}
