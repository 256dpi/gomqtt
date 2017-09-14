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

package packet

import "fmt"

// A Message bundles data that is published between brokers and clients.
type Message struct {
	// The Topic of the message.
	Topic string

	// The Payload of the message.
	Payload []byte

	// The QOS indicates the level of assurance for delivery.
	QOS byte

	// If the Retain flag is set to true, the server must store the message,
	// so that it can be delivered to future subscribers whose subscriptions
	// match its topic name.
	Retain bool
}

// String returns a string representation of the message.
func (m *Message) String() string {
	return fmt.Sprintf("<Message Topic=%q QOS=%d Retain=%t Payload=%v>",
		m.Topic, m.QOS, m.Retain, m.Payload)
}

// Copy returns a copy of the message.
func (m Message) Copy() *Message {
	return &m
}
