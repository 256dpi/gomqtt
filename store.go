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

package client

import (
	"sync"

	"github.com/gomqtt/packet"
)

// TODO: Maybe the store can be externalized and used by client, service and broker?

// Store is used to persists incoming or outgoing packets until they are
// successfully acknowledged by the other side.
type Store interface {
	// Open will open the store. Opening an already opened store must not
	// return an error. If clean is true the store gets also reset.
	Open(clean bool) error

	// Put will persist a packet to the store. An eventual existing packet with
	// the same id gets overwritten.
	Put(packet.Packet) error

	// Get will retrieve a packet from the store.
	Get(uint16) (packet.Packet, error)

	// Del will remove a packet from the store. Removing a nonexistent packet
	// must not return an error.
	Del(uint16) error

	// All will return all packets currently in the store.
	All() ([]packet.Packet, error)

	// Close will close the store. Closing an already closed store must not
	// return an error. Stores gets reset if clean was true during Open.
	Close() error
}

// MemoryStore organizes packets in memory.
type MemoryStore struct {
	clean bool
	store map[uint16]packet.Packet
	mutex sync.Mutex
}

// NewMemoryStore returns a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[uint16]packet.Packet),
	}
}

// Open opens the store and cleans it on request.
func (s *MemoryStore) Open(clean bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// cache clean setting
	s.clean = clean

	if s.clean {
		s.store = make(map[uint16]packet.Packet)
	}

	return nil
}

// Put will store the specified packet in the store.
func (s *MemoryStore) Put(pkt packet.Packet) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id, ok := packet.PacketID(pkt)
	if ok {
		s.store[id] = pkt
	}

	return nil
}

// Get will retrieve and return a packet by its packetId.
func (s *MemoryStore) Get(id uint16) (packet.Packet, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.store[id], nil
}

// Del will remove the a packet using its packetId.
func (s *MemoryStore) Del(id uint16) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.store, id)

	return nil
}

// All will return all stored packets.
func (s *MemoryStore) All() ([]packet.Packet, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	all := make([]packet.Packet, len(s.store))

	i := 0
	for _, pkt := range s.store {
		all[i] = pkt
		i++
	}

	return all, nil
}

// Close will close the store. If the store has been opened cleanly, Close will
// clean store again.
func (s *MemoryStore) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.clean {
		s.store = make(map[uint16]packet.Packet)
	}

	return nil
}
