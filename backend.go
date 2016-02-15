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
	"github.com/gomqtt/session"
	"github.com/gomqtt/packet"
)

type Backend interface {
	// GetSession returns the already stored session for the supplied id or creates
	// and returns a new one.
	GetSession(string) (session.Session, error)

	// Subscribe will subscribe the passed client to the specified topic and
	// begin to forward messages by calling the clients Publish method.
	Subscribe(client *Client, topic string) error

	// Unsubscribe will unsubscribe the passed client from the specified topic.
	Unsubscribe(client *Client, topic string) error

	// Remove will unsubscribe the passed client from previously subscribe topics.
	Remove(client *Client) error

	// Publish will forward the passed message to all other subscribed clients.
	Publish(client *Client, msg *packet.Message) error

	// StoreRetained retains the passed messages.
	StoreRetained(*Client, *packet.Message) error

	// RetrieveRetained will lookup all stored retained messages matching the
	// supplied topic.
	RetrieveRetained(*Client, string) ([]*packet.Message, error)
}

// TODO: missing offline subscriptions
