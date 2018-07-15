// Package session implements session objects to be used with MQTT clients and
// brokers.
package session

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/topic"
)

// Direction denotes a packets direction.
type Direction int

const (
	// Incoming packets are being received.
	Incoming Direction = iota

	// Outgoing packets are being be sent.
	Outgoing
)

// A MemorySession stores packets in memory.
type MemorySession struct {
	counter       *IDCounter
	incStore      *PacketStore
	outStore      *PacketStore
	subscriptions *topic.Tree
}

// NewMemorySession returns a new MemorySession.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter:       NewIDCounter(),
		incStore:      NewPacketStore(),
		outStore:      NewPacketStore(),
		subscriptions: topic.NewTree(),
	}
}

// NextID will return the next id for outgoing packets.
func (s *MemorySession) NextID() packet.ID {
	return s.counter.NextID()
}

// SavePacket will store a packet in the session. An eventual existing
// packet with the same id gets quietly overwritten.
func (s *MemorySession) SavePacket(dir Direction, pkt packet.GenericPacket) error {
	s.storeForDirection(dir).Save(pkt)
	return nil
}

// LookupPacket will retrieve a packet from the session using a packet id.
func (s *MemorySession) LookupPacket(dir Direction, id packet.ID) (packet.GenericPacket, error) {
	return s.storeForDirection(dir).Lookup(id), nil
}

// DeletePacket will remove a packet from the session. The method must not
// return an error if no packet with the specified id does exists.
func (s *MemorySession) DeletePacket(dir Direction, id packet.ID) error {
	s.storeForDirection(dir).Delete(id)
	return nil
}

// AllPackets will return all packets currently saved in the session.
func (s *MemorySession) AllPackets(dir Direction) ([]packet.GenericPacket, error) {
	return s.storeForDirection(dir).All(), nil
}

// SaveSubscription will store the subscription in the session. An eventual
// subscription with the same topic gets quietly overwritten.
func (s *MemorySession) SaveSubscription(sub packet.Subscription) error {
	s.subscriptions.Set(sub.Topic, sub)
	return nil
}

// LookupSubscription will match a topic against the stored subscriptions and
// eventually return the first found subscription.
func (s *MemorySession) LookupSubscription(topic string) (*packet.Subscription, error) {
	values := s.subscriptions.Match(topic)

	if len(values) > 0 {
		sub := values[0].(packet.Subscription)
		return &sub, nil
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
func (s *MemorySession) AllSubscriptions() ([]packet.Subscription, error) {
	var all []packet.Subscription

	for _, value := range s.subscriptions.All() {
		all = append(all, value.(packet.Subscription))
	}

	return all, nil
}

// Reset will completely reset the session.
func (s *MemorySession) Reset() error {
	s.counter.Reset()
	s.incStore.Reset()
	s.outStore.Reset()
	s.subscriptions.Reset()

	return nil
}

func (s *MemorySession) storeForDirection(dir Direction) *PacketStore {
	if dir == Incoming {
		return s.incStore
	} else if dir == Outgoing {
		return s.outStore
	}

	panic("unknown direction")
}
