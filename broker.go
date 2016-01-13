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

package broker

import (
	"github.com/gomqtt/transport"
)

type Broker struct {
	queueBackend QueueBackend
	retainedBackend RetainedBackend
	willBackend WillBackend
}

// New returns a new Broker.
func New(qb QueueBackend, rb RetainedBackend, wb WillBackend) *Broker {
	return &Broker{
		queueBackend: qb,
		retainedBackend: rb,
		willBackend: wb,
	}
}

// Handle handles a transport.Conn.
func (b *Broker) Handle(conn transport.Conn) {
	NewClient(b, conn)
}
