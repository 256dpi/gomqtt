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

// A store is used to persists incoming and outgoing packets until they are
// successfully acknowledged by the other side.
type Store interface {
	// Open will open the store.
	Open() error

	// Put will persist a packet to the store.
	Put(packet.Packet) error

	// Get will retrieve a packet from the store.
	Get(uint16) (packet.Packet, error)

	// Del will remove a packet from the store.
	Del(uint16) error

	// All will return all packets currently in the store.
	All() (error, []packet.Packet)

	// Reset will completely reset the store.
	Reset() error

	// Close will close the store.
	Close() error
}

type MemoryStore struct {
	store map[uint16]packet.Packet
	mutex sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[uint16]packet.Packet),
	}
}

func (s *MemoryStore) Open() error {
	return nil
}

func (s *MemoryStore) Put(pkt packet.Packet) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id, ok := packet.PacketID(pkt)
	if ok {
		s.store[id] = pkt
	}

	return nil
}

func (s *MemoryStore) Get(id uint16) (packet.Packet, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.store[id], nil
}

func (s *MemoryStore) Del(id uint16) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.store, id)

	return nil
}

func (s *MemoryStore) All() (error, []packet.Packet) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return nil, nil
}

func (s *MemoryStore) Reset() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store = make(map[uint16]packet.Packet)

	return nil
}

func (s *MemoryStore) Close() error {
	return nil
}
