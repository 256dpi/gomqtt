package client

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/tools"
)

// Direction denotes a packets direction.
type Direction int

const (
	// Outgoing packets are being be sent.
	Outgoing Direction = iota

	// Incoming packets are being received.
	Incoming
)

// A Session is used to persist incoming and outgoing packets.
type Session interface {
	// PacketID will return the next id for outgoing packets.
	PacketID() uint16

	// SavePacket will store a packet in the session. An eventual existing
	// packet with the same id gets quietly overwritten.
	SavePacket(Direction, packet.GenericPacket) error

	// LookupPacket will retrieve a packet from the session using a packet id.
	LookupPacket(dir Direction, id uint16) (packet.GenericPacket, error)

	// DeletePacket will remove a packet from the session. The method must not
	// return an error if no packet with the specified id does exists.
	DeletePacket(dir Direction, id uint16) error

	// AllPackets will return all packets currently saved in the session.
	AllPackets(Direction) ([]packet.GenericPacket, error)

	// Reset will completely reset the session.
	Reset() error
}

// A MemorySession stores packets in memory.
type MemorySession struct {
	counter  *tools.Counter
	incStore *tools.Store
	outStore *tools.Store
}

// NewMemorySession returns a new MemorySession.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter:  tools.NewCounter(),
		incStore: tools.NewStore(),
		outStore: tools.NewStore(),
	}
}

// PacketID will return the next id for outgoing packets.
func (s *MemorySession) PacketID() uint16 {
	return s.counter.Next()
}

// SavePacket will store a packet in the session. An eventual existing
// packet with the same id gets quietly overwritten.
func (s *MemorySession) SavePacket(dir Direction, pkt packet.GenericPacket) error {
	s.storeForDirection(dir).Save(pkt)
	return nil
}

// LookupPacket will retrieve a packet from the session using a packet id.
func (s *MemorySession) LookupPacket(dir Direction, id uint16) (packet.GenericPacket, error) {
	return s.storeForDirection(dir).Lookup(id), nil
}

// DeletePacket will remove a packet from the session. The method must not
// return an error if no packet with the specified id does exists.
func (s *MemorySession) DeletePacket(dir Direction, id uint16) error {
	s.storeForDirection(dir).Delete(id)
	return nil
}

// AllPackets will return all packets currently saved in the session.
func (s *MemorySession) AllPackets(dir Direction) ([]packet.GenericPacket, error) {
	return s.storeForDirection(dir).All(), nil
}

// Reset will completely reset the session.
func (s *MemorySession) Reset() error {
	s.counter.Reset()
	s.incStore.Reset()
	s.outStore.Reset()
	return nil
}

func (s *MemorySession) storeForDirection(dir Direction) *tools.Store {
	if dir == Incoming {
		return s.incStore
	} else if dir == Outgoing {
		return s.outStore
	}

	panic("unknown direction")
}
