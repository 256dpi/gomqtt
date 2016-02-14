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

	Subscribe(client *Client, topic string) error
	Unsubscribe(client *Client, topic string) error
	Remove(client *Client) error
	Publish(client *Client, msg *packet.Message) error
	//StoreRetained(*Client, string, []byte) error
	//RetrieveRetained(*Client, string) ([]byte, error)
}

// TODO: missing offline subscriptions
