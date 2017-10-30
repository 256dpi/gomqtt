package client

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

	list, err := session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))

	err = session.DeletePacket(incoming, 1)
	assert.NoError(t, err)

	pkt, err = session.LookupPacket(incoming, 1)
	assert.NoError(t, err)
	assert.Nil(t, pkt)

	list, err = session.AllPackets(incoming)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))

	err = session.SavePacket(outgoing, publish)
	assert.NoError(t, err)

	list, err = session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))

	err = session.Reset()
	assert.NoError(t, err)

	list, err = session.AllPackets(outgoing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))
}
