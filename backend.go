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
	"github.com/gomqtt/packet"
)

const (
	Outgoing = "out"
	Incoming = "in"
)

// Session is used to persist incoming/outgoing packets, subscriptions and the
// will.
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

	// SaveSubscription will store the subscription in the session. An eventual
	// subscription with the same topic gets quietly overwritten.
	SaveSubscription(sub *packet.Subscription) error

	// LookupSubscription will match a topic against the stored subscriptions and
	// eventually return the first found subscription.
	LookupSubscription(topic string) (*packet.Subscription, error)

	// DeleteSubscription will remove the subscription from the session. The
	// method must not return an error if no subscription with the specified
	// topic does exist.
	DeleteSubscription(topic string) error

	// AllSubscriptions will return all subscriptions currently saved in the
	// session.
	AllSubscriptions() ([]*packet.Subscription, error)

	// SaveWill will store the will message.
	SaveWill(msg *packet.Message) error

	// LookupWill will retrieve the will message.
	LookupWill() (*packet.Message, error)

	// ClearWill will remove the will message from the store.
	ClearWill() error

	// Reset will completely reset the session.
	Reset() error
}

type Backend interface {
	// GetSession returns the already stored session for the supplied id or creates
	// and returns a new one. If the supplied id has a zero length, a new
	// session is returned that is only valid once.
	GetSession(client *Client, id string) (Session, error)

	// Subscribe will subscribe the passed client to the specified topic and
	// begin to forward messages by calling the clients Publish method.
	// It will also return the stored retained messages matching the supplied
	// topic.
	Subscribe(client *Client, topic string) ([]*packet.Message, error)

	// Unsubscribe will unsubscribe the passed client from the specified topic.
	Unsubscribe(client *Client, topic string) error

	// Remove will unsubscribe the passed client from previously subscribe topics.
	Remove(client *Client) error

	// Publish will forward the passed message to all other subscribed clients.
	// It will also store the message if Retain is set to true. If the supplied
	//message has additionally a zero length payload, the backend removes the
	// currently retained message.
	Publish(client *Client, msg *packet.Message) error
}

// TODO: missing offline subscriptions
