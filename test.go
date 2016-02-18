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
	"testing"
	"math"

	"github.com/stretchr/testify/assert"
	"github.com/gomqtt/packet"
)

// BackendTest will test a Backend implementation. The test will test
// all methods using a fake consumer. The passed builder callback should always
// return a fresh instances of the Backend.
//
// Authentication: Its expected that the Backend allows the login "allow:allow".
func BackendTest(t *testing.T, builder func()(Backend)) {
	t.Log("Running Backend Authentication Test")
	backendAuthenticationTest(t, builder())

	t.Log("Running Backend Setup Test")
	backendSetupTest(t, builder())

	t.Log("Running Backend Basic Queuing Test")
	backendBasicQueuingTest(t, builder())

	// TODO: Offline Message Test?

	t.Log("Running Backend Retained Messages Test")
	backendRetainedMessagesTest(t, builder())
}

func backendAuthenticationTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	ok, err := backend.Authenticate(consumer, "allow", "allow")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = backend.Authenticate(consumer, "deny", "deny")
	assert.False(t, ok)
	assert.NoError(t, err)
}

func backendSetupTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	// has id and clean=false

	session1, resumed, err := backend.Setup(consumer, "foo", false)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session1)

	session2, resumed, err := backend.Setup(consumer, "foo", false)
	assert.NoError(t, err)
	assert.True(t, resumed)
	assert.True(t, session1 == session2)

	// has id and clean=true

	session3, resumed, err := backend.Setup(consumer, "foo", true)
	assert.NoError(t, err)
	assert.True(t, resumed)
	assert.True(t, session1 == session3)

	// has other id and clean=false

	session4, resumed, err := backend.Setup(consumer, "bar", false)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.True(t, session4 != session1)
	assert.True(t, session4 != session2)
	assert.True(t, session4 != session3)

	// has no id and clean=true

	session5, resumed, err := backend.Setup(consumer, "", true)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session5)

	session6, resumed, err := backend.Setup(consumer, "", true)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session5)
	assert.True(t, session5 != session6)
}

func backendBasicQueuingTest(t *testing.T, backend Backend) {
	consumer1 := newFakeConsumer()
	consumer2 := newFakeConsumer()

	msg := &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
	}

	// setup both consumers

	_, _, err := backend.Setup(consumer1, "consumer1", true)
	assert.NoError(t, err)

	_, _, err = backend.Setup(consumer2, "consumer2", true)
	assert.NoError(t, err)

	// subscribe both consumers

	msgs, err := backend.Subscribe(consumer1, "test")
	assert.Nil(t, msgs)
	assert.NoError(t, err)

	msgs, err = backend.Subscribe(consumer2, "test")
	assert.Nil(t, msgs)
	assert.NoError(t, err)

	// publish message

	err = backend.Publish(consumer1, msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, consumer1.in[0])
	assert.Equal(t, msg, consumer2.in[0])

	// unsubscribe one consumer

	err = backend.Unsubscribe(consumer2, "test")
	assert.NoError(t, err)

	// publish another message

	err = backend.Publish(consumer1, msg)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(consumer1.in))
	assert.Equal(t, 1, len(consumer2.in))

	// terminate both consumers

	err = backend.Terminate(consumer1)
	assert.NoError(t, err)

	err = backend.Terminate(consumer2)
	assert.NoError(t, err)
}

func backendRetainedMessagesTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	msg1 := &packet.Message{
		Topic:   "foo",
		Payload: []byte("bar"),
		QOS:     1,
		Retain:  true,
	}

	msg2 := &packet.Message{
		Topic:   "foo/bar",
		Payload: []byte("bar"),
		QOS:     1,
		Retain:  true,
	}

	msg3 := &packet.Message{
		Topic:   "foo",
		Payload: []byte("bar"),
		QOS:     2,
		Retain:  true,
	}

	msg4 := &packet.Message{
		Topic:  "foo",
		QOS:    1,
		Retain: true,
	}

	// should be empty
	msgs, err := backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)

	err = backend.Publish(consumer, msg1)
	assert.NoError(t, err)

	// should have one
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg1, msgs[0])

	err = backend.Publish(consumer, msg2)
	assert.NoError(t, err)

	// should have two
	msgs, err = backend.Subscribe(consumer, "#")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(msgs))

	err = backend.Publish(consumer, msg3)
	assert.NoError(t, err)

	// should have another
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg3, msgs[0])

	err = backend.Publish(consumer, msg4)
	assert.NoError(t, err)

	// should have none
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}

// SessionTest will test a Session implementation. The passed builder callback
// should always return a fresh instances of the Session.
func SessionTest(t *testing.T, builder func()(Session)) {
	t.Log("Running Session Packet ID Test")
	sessionPacketIDTest(t, builder())

	t.Log("Running Packet Store Test")
	sessionPacketStoreTest(t, builder())

	t.Log("Running Subscription Store Test")
	sessionSubscriptionStoreTest(t, builder())

	t.Log("Running Will Store Test")
	sessionWillStoreTest(t, builder())
}

func sessionPacketIDTest(t *testing.T, session Session) {
	assert.Equal(t, uint16(1), session.PacketID())
	assert.Equal(t, uint16(2), session.PacketID())

	for i := 0; i < math.MaxUint16-3; i++ {
		session.PacketID()
	}

	assert.Equal(t, uint16(math.MaxUint16), session.PacketID())
	assert.Equal(t, uint16(1), session.PacketID())

	err := session.Reset()
	assert.NoError(t, err)

	assert.Equal(t, uint16(1), session.PacketID())
}

func sessionPacketStoreTest(t *testing.T, session Session) {
	publish := packet.NewPublishPacket()
	publish.PacketID = 1

	pkt, err := session.LookupPacket(incoming, 1)
	assert.NoError(t, err)
	assert.Nil(t, pkt)

	err = session.SavePacket(incoming, publish)
	assert.NoError(t, err)

	pkt, err = session.LookupPacket(incoming, 1)
	assert.NoError(t, err)
	assert.Equal(t, publish, pkt)

	pkts, err := session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pkts))

	err = session.DeletePacket(incoming, 1)
	assert.NoError(t, err)

	pkt, err = session.LookupPacket(incoming, 1)
	assert.NoError(t, err)
	assert.Nil(t, pkt)

	pkts, err = session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pkts))

	err = session.SavePacket(outgoing, publish)
	assert.NoError(t, err)

	pkts, err = session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pkts))

	err = session.Reset()
	assert.NoError(t, err)

	pkts, err = session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pkts))
}

func sessionSubscriptionStoreTest(t *testing.T, session Session) {
	subscription := &packet.Subscription{
		Topic: "+",
		QOS:   1,
	}

	subs, err := session.AllSubscriptions()
	assert.Equal(t, 0, len(subs))

	sub, err := session.LookupSubscription("foo")
	assert.Nil(t, sub)
	assert.NoError(t, err)

	err = session.SaveSubscription(subscription)
	assert.NoError(t, err)

	sub, err = session.LookupSubscription("foo")
	assert.Equal(t, subscription, sub)
	assert.NoError(t, err)

	subs, err = session.AllSubscriptions()
	assert.Equal(t, 1, len(subs))

	err = session.DeleteSubscription("+")
	assert.NoError(t, err)

	sub, err = session.LookupSubscription("foo")
	assert.Nil(t, sub)
	assert.NoError(t, err)

	subs, err = session.AllSubscriptions()
	assert.Equal(t, 0, len(subs))

	err = session.SaveSubscription(subscription)
	assert.NoError(t, err)

	subs, err = session.AllSubscriptions()
	assert.Equal(t, 1, len(subs))

	err = session.Reset()
	assert.NoError(t, err)

	subs, err = session.AllSubscriptions()
	assert.Equal(t, 0, len(subs))
}

func sessionWillStoreTest(t *testing.T, session Session) {
	theWill := &packet.Message{"test", []byte("test"), 0, false}

	will, err := session.LookupWill()
	assert.Nil(t, will)
	assert.NoError(t, err)

	err = session.SaveWill(theWill)
	assert.NoError(t, err)

	will, err = session.LookupWill()
	assert.Equal(t, theWill, will)
	assert.NoError(t, err)

	err = session.ClearWill()
	assert.NoError(t, err)

	will, err = session.LookupWill()
	assert.Nil(t, will)
	assert.NoError(t, err)

	err = session.SaveWill(theWill)
	assert.NoError(t, err)

	will, err = session.LookupWill()
	assert.Equal(t, theWill, will)
	assert.NoError(t, err)

	err = session.Reset()
	assert.NoError(t, err)

	will, err = session.LookupWill()
	assert.Nil(t, will)
	assert.NoError(t, err)
}
