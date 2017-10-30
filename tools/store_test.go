package tools

import (
	"testing"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	store := NewStore()

	publish := packet.NewPublishPacket()
	publish.PacketID = 1

	pkt := store.Lookup(1)
	assert.Nil(t, pkt)

	store.Save(publish)

	pkt = store.Lookup(1)
	assert.Equal(t, publish, pkt)

	pkts := store.All()
	assert.Equal(t, 1, len(pkts))

	store.Delete(1)

	pkt = store.Lookup(1)
	assert.Nil(t, pkt)

	pkts = store.All()
	assert.Equal(t, 0, len(pkts))
}
