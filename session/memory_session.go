// Package session implements session objects to be used with MQTT clients and
// brokers.
package session

import (
	"github.com/256dpi/gomqtt/packet"
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
	counter  *IDCounter
	incStore *PacketStore
	outStore *PacketStore
}

// NewMemorySession returns a new MemorySession.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter:  NewIDCounter(),
		incStore: NewPacketStore(),
		outStore: NewPacketStore(),
	}
}

// NextID will return the next id for outgoing packets.
func (s *MemorySession) NextID() packet.ID {
	return s.counter.NextID()
}

// SavePacket will store a packet in the session. An eventual existing
// packet with the same id gets quietly overwritten.
func (s *MemorySession) SavePacket(dir Direction, pkt packet.Generic) error {
	s.storeForDirection(dir).Save(pkt)
	return nil
}

// LookupPacket will retrieve a packet from the session using a packet id.
func (s *MemorySession) LookupPacket(dir Direction, id packet.ID) (packet.Generic, error) {
	return s.storeForDirection(dir).Lookup(id), nil
}

// DeletePacket will remove a packet from the session. The method must not
// return an error if no packet with the specified id does exists.
func (s *MemorySession) DeletePacket(dir Direction, id packet.ID) error {
	s.storeForDirection(dir).Delete(id)
	return nil
}

// AllPackets will return all packets currently saved in the session.
func (s *MemorySession) AllPackets(dir Direction) ([]packet.Generic, error) {
	return s.storeForDirection(dir).All(), nil
}

// Reset will completely reset the session.
func (s *MemorySession) Reset() error {
	s.counter.Reset()
	s.incStore.Reset()
	s.outStore.Reset()

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
