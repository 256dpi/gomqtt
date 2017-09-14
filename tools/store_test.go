package tools

import (
	"testing"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	store := NewStore()

	publish := packet.NewPublishPacket()
	publish.PacketID = 1

	pkt := store.Lookup("in", 1)
	assert.Nil(t, pkt)

	store.Save("in", publish)

	pkt = store.Lookup("in", 1)
	assert.Equal(t, publish, pkt)

	pkts := store.All("in")
	assert.Equal(t, 1, len(pkts))

	store.Delete("in", 1)

	pkt = store.Lookup("in", 1)
	assert.Nil(t, pkt)

	pkts = store.All("in")
	assert.Equal(t, 0, len(pkts))

	store.Save("out", publish)

	pkts = store.All("out")
	assert.Equal(t, 1, len(pkts))

	store.Reset()

	pkts = store.All("out")
	assert.Equal(t, 0, len(pkts))
}
