# gomqtt/broker

[![Build Status](https://travis-ci.org/gomqtt/broker.svg?branch=master)](https://travis-ci.org/gomqtt/broker)
[![Coverage Status](https://coveralls.io/repos/github/gomqtt/broker/badge.svg?branch=master)](https://coveralls.io/github/gomqtt/broker?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/broker?status.svg)](http://godoc.org/github.com/gomqtt/broker)
[![Release](https://img.shields.io/github/release/gomqtt/broker.svg)](https://github.com/gomqtt/broker/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomqtt/broker)](https://goreportcard.com/report/github.com/gomqtt/broker)

**Package broker provides an extensible [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) broker implementation.**

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/broker
```

## Usage

```go
// A Backend provides effective queuing functionality to a Broker and its Clients.
type Backend interface {
	// Authenticate should authenticate the client using the user and password
	// values and return true if the client is eligible to continue or false
	// when the broker should terminate the connection.
	Authenticate(client *Client, user, password string) (bool, error)

	// Setup is called when a new client comes online and is successfully
	// authenticated. Setup should return the already stored session for the
	// supplied id or create and return a new one. If the supplied id has a zero
	// length, a new temporary session should returned that is not stored
	// further. The backend may also close any existing clients that use the
	// same client id.
	//
	// Note: In this call the Backend may also allocate other resources and
	// setup the client for further usage as the broker will acknowledge the
	// connection when the call returns.
	Setup(client *Client, id string) (Session, bool, error)

	// QueueOffline is called after the clients stored subscriptions have been
	// resubscribed. It should be used to trigger a background process that
	// forwards all missed messages.
	QueueOffline(client *Client) error

	// Subscribe should subscribe the passed client to the specified topic and
	// call Publish with any incoming messages.
	Subscribe(client *Client, topic string) error

	// Unsubscribe should unsubscribe the passed client from the specified topic.
	Unsubscribe(client *Client, topic string) error

	// StoreRetained should store the specified message.
	StoreRetained(client *Client, msg *packet.Message) error

	// ClearRetained should remove the stored messages for the given topic.
	ClearRetained(client *Client, topic string) error

	// QueueRetained is called after acknowledging a subscription and should be
	// used to trigger a background process that forwards all retained messages.
	QueueRetained(client *Client, topic string) error

	// Publish should forward the passed message to all other clients that hold
	// a subscription that matches the messages topic. It should also add the
	// message to all sessions that have a matching offline subscription.
	Publish(client *Client, msg *packet.Message) error

	// Terminate is called when the client goes offline. Terminate should
	// unsubscribe the passed client from all previously subscribed topics. The
	// backend may also convert a clients subscriptions to offline subscriptions.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that client as the broker will close the connection when the call
	// returns.
	Terminate(client *Client) error
}

// A Session is used to persist incoming/outgoing packets, subscriptions and the
// will.
type Session interface {
	// PacketID should return the next id for outgoing packets.
	PacketID() uint16

	// SavePacket should store a packet in the session. An eventual existing
	// packet with the same id should be quietly overwritten.
	SavePacket(direction string, pkt packet.Packet) error

	// LookupPacket should retrieve a packet from the session using the packet id.
	LookupPacket(direction string, id uint16) (packet.Packet, error)

	// DeletePacket should remove a packet from the session. The method should
	// not return an error if no packet with the specified id does exists.
	DeletePacket(direction string, id uint16) error

	// AllPackets should return all packets currently saved in the session. This
	// method is used to resend stored packets when the session is resumed.
	AllPackets(direction string) ([]packet.Packet, error)

	// SaveSubscription should store the subscription in the session. An eventual
	// subscription with the same topic should be quietly overwritten.
	SaveSubscription(sub *packet.Subscription) error

	// LookupSubscription should match a topic against the stored subscriptions
	// and eventually return the first found subscription.
	LookupSubscription(topic string) (*packet.Subscription, error)

	// DeleteSubscription should remove the subscription from the session. The
	// method should not return an error if no subscription with the specified
	// topic does exist.
	DeleteSubscription(topic string) error

	// AllSubscriptions should return all subscriptions currently saved in the
	// session. This method is used to restore a clients subscriptions when the
	// session is resumed.
	AllSubscriptions() ([]*packet.Subscription, error)

	// SaveWill should store the will message.
	SaveWill(msg *packet.Message) error

	// LookupWill should retrieve the will message.
	LookupWill() (*packet.Message, error)

	// ClearWill should remove the will message from the store.
	ClearWill() error

	// Reset should completely reset the session.
	Reset() error
}
```
