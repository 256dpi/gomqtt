package session

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
)

// PacketStore is a goroutine safe packet store.
type PacketStore struct {
	packets map[packet.ID]packet.Generic
	mutex   sync.RWMutex
}

// NewPacketStore returns a new PacketStore.
func NewPacketStore() *PacketStore {
	return &PacketStore{
		packets: map[packet.ID]packet.Generic{},
	}
}

// NewPacketStoreWithPackets returns a new PacketStore with the provided packets.
func NewPacketStoreWithPackets(packets []packet.Generic) *PacketStore {
	// create store
	store := NewPacketStore()

	// add packets
	for _, pkt := range packets {
		store.Save(pkt)
	}

	return store
}

// Save will store a packet in the store. An eventual existing packet with the
// same id gets quietly overwritten.
func (s *PacketStore) Save(pkt packet.Generic) {
	// acquire mutex
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// get packet id
	id, ok := packet.GetID(pkt)
	if !ok {
		panic("failed to get packet id")
	}

	// store packet
	s.packets[id] = pkt
}

// Lookup will retrieve a packet from the store.
func (s *PacketStore) Lookup(id packet.ID) packet.Generic {
	// acquire mutex
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// get packet
	return s.packets[id]
}

// Delete will remove a packet from the store.
func (s *PacketStore) Delete(id packet.ID) {
	// acquire mutex
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// delete packet
	delete(s.packets, id)
}

// All will return all packets currently saved in the store.
func (s *PacketStore) All() []packet.Generic {
	// acquire mutex
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// collect packets
	all := make([]packet.Generic, 0, len(s.packets))
	for _, pkt := range s.packets {
		all = append(all, pkt)
	}

	return all
}

// Reset will reset the store.
func (s *PacketStore) Reset() {
	// acquire mutex
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// reset packets
	s.packets = map[packet.ID]packet.Generic{}
}
