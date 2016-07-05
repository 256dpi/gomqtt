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
	"sync"
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

	closing   bool
	clients   []*Client
	mutex     sync.Mutex
	waitGroup sync.WaitGroup
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
		clients:        make([]*Client, 0),
	}
}

// Handle takes over responsibility and handles a transport.Conn. It returns
// false if the broker is closing and the connection has been closed.
func (b *Broker) Handle(conn transport.Conn) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// check conn
	if conn == nil {
		panic("passed conn is nil")
	}

	// close conn immediately when closing
	if b.closing {
		conn.Close()
		return false
	}

	// handle client
	newClient(b, conn)

	return true
}

// Clients returns a current list of connected clients.
func (b *Broker) Clients() []*Client {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// copy list
	clients := make([]*Client, len(b.clients))
	copy(clients, b.clients)

	return clients
}

// Close will stop handling incoming connections and close all current clients.
// The call will block until all clients are properly closed.
func (b *Broker) Close() {
	b.mutex.Lock()

	// set closing
	b.closing = true

	// close all clients
	for _, client := range b.clients {
		client.Close(false)
	}

	// wait for all clients to close
	b.mutex.Unlock()
	b.waitGroup.Wait()
}

// clients call add to add themselves to the list
func (b *Broker) add(client *Client) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// add client
	b.clients = append(b.clients, client)

	// increment wait group
	b.waitGroup.Add(1)
}

// clients call remove when closed to remove themselves from the list
func (b *Broker) remove(client *Client) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// get index of client
	index := 0
	for i, c := range b.clients {
		if c == client {
			index = i
			break
		}
	}

	// remove client from list
	// https://github.com/golang/go/wiki/SliceTricks
	b.clients[index] = b.clients[len(b.clients)-1]
	b.clients[len(b.clients)-1] = nil
	b.clients = b.clients[:len(b.clients)-1]

	// decrement wait group
	b.waitGroup.Add(-1)
}
