package broker

import (
	"math"
	"testing"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestMemorySessionPacketID(t *testing.T) {
	session := NewMemorySession()

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

func TestMemorySessionPacketStore(t *testing.T) {
	session := NewMemorySession()

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

func TestMemorySessionSubscriptionStore(t *testing.T) {
	session := NewMemorySession()

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
