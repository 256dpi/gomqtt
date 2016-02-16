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

type MemorySession struct {
	counter       *tools.Counter
	store         *tools.Store
	subscriptions *tools.Tree

	will      *packet.Message
	willMutex sync.Mutex
}

func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter:       tools.NewCounter(),
		store:         tools.NewStore(),
		subscriptions: tools.NewTree(),
	}
}

// PacketID will return the next id for outgoing packets.
func (s *MemorySession) PacketID() uint16 {
	return s.counter.Next()
}

// SavePacket will store a packet in the session. An eventual existing
// packet with the same id gets quietly overwritten.
func (s *MemorySession) SavePacket(direction string, pkt packet.Packet) error {
	s.store.Save(direction, pkt)
	return nil
}

// LookupPacket will retrieve a packet from the session using a packet id.
func (s *MemorySession) LookupPacket(direction string, id uint16) (packet.Packet, error) {
	return s.store.Lookup(direction, id), nil
}

// DeletePacket will remove a packet from the session. The method must not
// return an error if no packet with the specified id does exists.
func (s *MemorySession) DeletePacket(direction string, id uint16) error {
	s.store.Delete(direction, id)
	return nil
}

// AllPackets will return all packets currently saved in the session.
func (s *MemorySession) AllPackets(direction string) ([]packet.Packet, error) {
	return s.store.All(direction), nil
}

// SaveSubscription will store the subscription in the session. An eventual
// subscription with the same topic gets quietly overwritten.
func (s *MemorySession) SaveSubscription(sub *packet.Subscription) error {
	s.subscriptions.Set(sub.Topic, sub)
	return nil
}

// LookupSubscription will match a topic against the stored subscriptions and
// eventually return the first found subscription.
func (s *MemorySession) LookupSubscription(topic string) (*packet.Subscription, error) {
	values := s.subscriptions.Match(topic)

	if len(values) > 0 {
		if sub, ok := values[0].(*packet.Subscription); ok {
			return sub, nil
		}
	}

	return nil, nil
}

// DeleteSubscription will remove the subscription from the session. The
// method must not return an error if no subscription with the specified
// topic does exist.
func (s *MemorySession) DeleteSubscription(topic string) error {
	s.subscriptions.Empty(topic)
	return nil
}

// AllSubscriptions will return all subscriptions currently saved in the session.
func (s *MemorySession) AllSubscriptions() ([]*packet.Subscription, error) {
	var all []*packet.Subscription

	for _, value := range s.subscriptions.All() {
		if sub, ok := value.(*packet.Subscription); ok {
			all = append(all, sub)
		}
	}

	return all, nil
}

// SaveWill will store the will message.
func (s *MemorySession) SaveWill(newWill *packet.Message) error {
	s.willMutex.Lock()
	defer s.willMutex.Unlock()

	s.will = newWill

	return nil
}

// LookupWill will retrieve the will message.
func (s *MemorySession) LookupWill() (*packet.Message, error) {
	s.willMutex.Lock()
	defer s.willMutex.Unlock()

	return s.will, nil
}

// ClearWill will remove the will message from the store.
func (s *MemorySession) ClearWill() error {
	s.willMutex.Lock()
	defer s.willMutex.Unlock()

	s.will = nil

	return nil
}

// Reset will completely reset the session.
func (s *MemorySession) Reset() error {
	s.counter.Reset()
	s.store.Reset()
	s.subscriptions.Reset()
	s.ClearWill()

	return nil
}

type MemoryBackend struct {
	queue    *tools.Tree
	retained *tools.Tree

	sessions      map[string]*MemorySession
	sessionsMutex sync.Mutex
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		queue:    tools.NewTree(),
		retained: tools.NewTree(),
		sessions: make(map[string]*MemorySession),
	}
}

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
