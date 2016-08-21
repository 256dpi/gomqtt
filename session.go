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

const (
	outgoing = "out"
	incoming = "in"
)

// A Session is used to persist incoming/outgoing packets, subscriptions and the
// will.
type Session interface {
	// PacketID should return the next id for outgoing packets.
	PacketID() uint16

	// SavePacket should store a packet in the session. An eventual existing
	// packet with the same id should be quietly overwritten.
	SavePacket(direction string, pkt packet.Packet) error

	// LookupPacket should retrieve a packet from the session using the packet id.
	LookupPacket(direction string, id uint16) (packet.Packet, error)

	// DeletePacket should remove a packet from the session. The method should
	// not return an error if no packet with the specified id does exists.
	DeletePacket(direction string, id uint16) error

	// AllPackets should return all packets currently saved in the session. This
	// method is used to resend stored packets when the session is resumed.
	AllPackets(direction string) ([]packet.Packet, error)

	// SaveSubscription should store the subscription in the session. An eventual
	// subscription with the same topic should be quietly overwritten.
	SaveSubscription(sub *packet.Subscription) error

	// LookupSubscription should match a topic against the stored subscriptions
	// and eventually return the first found subscription.
	LookupSubscription(topic string) (*packet.Subscription, error)

	// DeleteSubscription should remove the subscription from the session. The
	// method should not return an error if no subscription with the specified
	// topic does exist.
	DeleteSubscription(topic string) error

	// AllSubscriptions should return all subscriptions currently saved in the
	// session. This method is used to restore a clients subscriptions when the
	// session is resumed.
	AllSubscriptions() ([]*packet.Subscription, error)

	// SaveWill should store the will message.
	SaveWill(msg *packet.Message) error

	// LookupWill should retrieve the will message.
	LookupWill() (*packet.Message, error)

	// ClearWill should remove the will message from the store.
	ClearWill() error

	// Reset should completely reset the session.
	Reset() error
}

// A MemorySession stores packets, subscriptions and the will in memory.
type MemorySession struct {
	counter       *tools.Counter
	store         *tools.Store
	subscriptions *tools.Tree
	offlineStore  *tools.Queue

	will      *packet.Message
	willMutex sync.Mutex
}

// NewMemorySession returns a new MemorySession.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter:       tools.NewCounter(),
		store:         tools.NewStore(),
		subscriptions: tools.NewTree(),
		offlineStore:  tools.NewQueue(100),
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

// DeletePacket will remove a packet from the session. The method will not
// return an error if no packet with the specified id exists.
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
		return values[0].(*packet.Subscription), nil
	}

	return nil, nil
}

// DeleteSubscription will remove the subscription from the session. The
// method will not return an error if no subscription with the specified
// topic does exist.
func (s *MemorySession) DeleteSubscription(topic string) error {
	s.subscriptions.Empty(topic)
	return nil
}

// AllSubscriptions will return all subscriptions currently saved in the session.
func (s *MemorySession) AllSubscriptions() ([]*packet.Subscription, error) {
	var all []*packet.Subscription

	for _, value := range s.subscriptions.All() {
		all = append(all, value.(*packet.Subscription))
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

// called by the backend to queue an offline message
func (s *MemorySession) queue(msg *packet.Message) {
	s.offlineStore.Push(msg)
}

// called by the backend to retrieve all offline messages
func (s *MemorySession) nextMissed() *packet.Message {
	return s.offlineStore.Pop()
}
