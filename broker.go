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
	"time"

	"github.com/gomqtt/transport"
)

// LogEvent are received by a Logger.
type LogEvent int

const (
	// NewConnectionLogEvent is emitted when a client comes online. The third
	// parameter will be nil.
	NewConnectionLogEvent LogEvent = iota

	// PacketReceivedLogEvent is emitted when a packet has been received. The
	// third parameter will be the received packet.
	PacketReceivedLogEvent

	// MessagePublishedLogEvent is emitted after a message has been published.
	// The third parameter will be the published message.
	MessagePublishedLogEvent

	// MessageForwardedLogEvent is emitted after a message has been forwarded.
	// The third parameter will be the forwarded message.
	MessageForwardedLogEvent

	// PacketSentLogEvent is emitted when a packet has been sent. The third
	// parameter will be the sent packet.
	PacketSentLogEvent

	// LostConnectionLogEvent is emitted when the connection has been terminated.
	// The third parameter will be nil.
	LostConnectionLogEvent

	// ErrorLogEvent is emitted when an error occurs. The third parameter will be
	// the caught error.
	ErrorLogEvent
)

// The Logger callback handles incoming log messages.
type Logger func(LogEvent, *Client, interface{})

// The Broker handles incoming connections and connects them to the backend.
type Broker struct {
	Backend Backend
	Logger  Logger

	ConnectTimeout time.Duration
}

// New returns a new Broker with a basic MemoryBackend.
func New() *Broker {
	return NewWithBackend(NewMemoryBackend())
}

// NewWithBackend returns a new Broker with a custom Backend.
func NewWithBackend(backend Backend) *Broker {
	return &Broker{
		Backend:        backend,
		ConnectTimeout: 10 * time.Second,
	}
}

// Handle takes over responsibility and handles a transport.Conn.
func (b *Broker) Handle(conn transport.Conn) {
	if conn == nil {
		panic("passed conn is nil")
	}

	newClient(b, conn)
}
