package session

import (
	"math"
	"testing"

	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

func TestMemorySessionNextID(t *testing.T) {
	session := NewMemorySession()

	assert.Equal(t, packet.ID(1), session.NextID())
	assert.Equal(t, packet.ID(2), session.NextID())

	for i := 0; i < math.MaxUint16-3; i++ {
		session.NextID()
	}

	assert.Equal(t, packet.ID(math.MaxUint16), session.NextID())
	assert.Equal(t, packet.ID(1), session.NextID())

	err := session.Reset()
	assert.NoError(t, err)

	assert.Equal(t, packet.ID(1), session.NextID())
}

func TestMemorySessionPacketStore(t *testing.T) {
	session := NewMemorySession()

	publish := packet.NewPublishPacket()
	publish.ID = 1

	pkt, err := session.LookupPacket(Incoming, 1)
	assert.NoError(t, err)
	assert.Nil(t, pkt)

	err = session.SavePacket(Incoming, publish)
	assert.NoError(t, err)

	pkt, err = session.LookupPacket(Incoming, 1)
	assert.NoError(t, err)
	assert.Equal(t, publish, pkt)

	list, err := session.AllPackets(Incoming)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))

	err = session.DeletePacket(Incoming, 1)
	assert.NoError(t, err)

	pkt, err = session.LookupPacket(Incoming, 1)
	assert.NoError(t, err)
	assert.Nil(t, pkt)

	list, err = session.AllPackets(Incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))

	err = session.SavePacket(Outgoing, publish)
	assert.NoError(t, err)

	list, err = session.AllPackets(Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))

	err = session.Reset()
	assert.NoError(t, err)

	list, err = session.AllPackets(Outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))
}

func TestMemorySessionSubscriptionStore(t *testing.T) {
	session := NewMemorySession()

	subscription := packet.Subscription{
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
	assert.Equal(t, &subscription, sub)
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

func TestMemorySessionWillStore(t *testing.T) {
	session := NewMemorySession()

	theWill := &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     0,
		Retain:  false,
	}

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
