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
)

// TODO: Maybe the store can be externalized and used by client, service and broker?

const(
	Incoming string = "in"
	Outgoing string = "out"
)

// Store is used to persists incoming or outgoing packets until they are
// successfully acknowledged by the other side.
type Store interface {
	// Put will persist a packet to the store. An eventual existing packet with
	// the same id gets overwritten.
	Put(string, packet.Packet) error

	// Get will retrieve a packet from the store.
	Get(string, uint16) (packet.Packet, error)

	// Del will remove a packet from the store. Removing a nonexistent packet
	// must not return an error.
	Del(string, uint16) error

	// All will return all packets currently in the store.
	All(string) ([]packet.Packet, error)

	// Reset will wipe all packets currently stored.
	Reset() error
}
