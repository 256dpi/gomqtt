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

import "github.com/gomqtt/packet"

// A store is used to persists incoming and outgoing packets until they are
// successfully acknowledged by the other side.
type Store interface {
	// Open will open the store.
	Open() error

	// Put will persist a packet to the store.
	Put(*packet.Packet) error

	// Dell will remove a packet from the store.
	Del(*packet.Packet) error

	// All will return all packets currently in the store.
	All() (error, []packet.Packet)

	// Close will close the store.
	Close() error

	// Reset will completely reset the store.
	Reset() error
}
