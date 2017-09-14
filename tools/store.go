package tools

import (
	"strings"
	"sync"

	"github.com/gomqtt/packet"
)

// The Store is a thread-safe packet store.
type Store struct {
	packets map[string]packet.Packet
	mutex   sync.Mutex
}

// NewStore returns a new Store.
func NewStore() *Store {
	return &Store{
		packets: make(map[string]packet.Packet),
	}
}

// Save will store a packet in the store. An eventual existing packet with the
// same id gets quietly overwritten.
func (s *Store) Save(prefix string, pkt packet.Packet) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id, ok := packet.PacketID(pkt)
	if ok {
		s.packets[s.key(prefix, id)] = pkt
	}
}

// Lookup will retrieve a packet from the store.
func (s *Store) Lookup(prefix string, id uint16) packet.Packet {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.packets[s.key(prefix, id)]
}

// Delete will remove a packet from the store.
func (s *Store) Delete(prefix string, id uint16) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.packets, s.key(prefix, id))
}

// All will return all packets currently saved in the store.
func (s *Store) All(prefix string) []packet.Packet {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var all []packet.Packet

	for key, pkt := range s.packets {
		if strings.HasPrefix(key, prefix) {
			all = append(all, pkt)
		}
	}

	return all
}

// Reset will reset the store.
func (s *Store) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.packets = make(map[string]packet.Packet)
}

// key generates a key from a prefix and a id
func (s *Store) key(prefix string, id uint16) string {
	return prefix + "-" + string(id)
}
