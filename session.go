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
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
)

const (
	outgoing = "out"
	incoming = "in"
)

// A Session is used to persist incoming and outgoing packets.
type Session interface {
	// PacketID will return the next id for outgoing packets.
	PacketID() uint16

	// SavePacket will store a packet in the session. An eventual existing
	// packet with the same id gets quietly overwritten.
	SavePacket(direction string, pkt packet.Packet) error

	// LookupPacket will retrieve a packet from the session using a packet id.
	LookupPacket(direction string, id uint16) (packet.Packet, error)

	// DeletePacket will remove a packet from the session. The method must not
	// return an error if no packet with the specified id does exists.
	DeletePacket(direction string, id uint16) error

	// AllPackets will return all packets currently saved in the session.
	AllPackets(direction string) ([]packet.Packet, error)

	// Reset will completely reset the session.
	Reset() error
}

// A MemorySession stores packets in memory.
type MemorySession struct {
	counter *tools.Counter
	store   *tools.Store
}

// NewMemorySession returns a new MemorySession.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter: tools.NewCounter(),
		store:   tools.NewStore(),
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

// Reset will completely reset the session.
func (s *MemorySession) Reset() error {
	s.counter.Reset()
	s.store.Reset()
	return nil
}
