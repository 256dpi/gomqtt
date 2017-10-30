package tools

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
)

// The Store is a thread-safe packet store.
type Store struct {
	packets map[packet.ID]packet.GenericPacket
	mutex   sync.Mutex
}

// NewStore returns a new Store.
func NewStore() *Store {
	return &Store{
		packets: make(map[packet.ID]packet.GenericPacket),
	}
}

// Save will store a packet in the store. An eventual existing packet with the
// same id gets quietly overwritten.
func (s *Store) Save(pkt packet.GenericPacket) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id, ok := packet.GetID(pkt)
	if ok {
		s.packets[id] = pkt
	}
}

// Lookup will retrieve a packet from the store.
func (s *Store) Lookup(id packet.ID) packet.GenericPacket {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.packets[id]
}

// Delete will remove a packet from the store.
func (s *Store) Delete(id packet.ID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.packets, id)
}

// All will return all packets currently saved in the store.
func (s *Store) All() []packet.GenericPacket {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var all []packet.GenericPacket

	for _, pkt := range s.packets {
		all = append(all, pkt)
	}

	return all
}

// Reset will reset the store.
func (s *Store) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.packets = make(map[packet.ID]packet.GenericPacket)
}
