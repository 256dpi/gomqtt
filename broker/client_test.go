package broker

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

type testMemoryBackend struct {
	MemoryBackend

	packets []packet.Generic
}

func (b *testMemoryBackend) Setup(client *Client, id string, clean bool) (Session, bool, error) {
	client.PacketCallback = func(pkt packet.Generic) error {
		b.packets = append(b.packets, pkt)
		return nil
	}

	return b.MemoryBackend.Setup(client, id, clean)
}

func TestClientPacketCallback(t *testing.T) {
	backend := &testMemoryBackend{
		MemoryBackend: *NewMemoryBackend(),
	}

	port, quit, done := Run(NewEngine(backend), "tcp")

	options := client.NewConfig("tcp://localhost:" + port)

	client1 := client.New()

	cf, err := client1.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))

	sf, err := client1.Subscribe("cool", 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))

	err = client1.Disconnect()
	assert.NoError(t, err)

	ret := backend.Close(5 * time.Second)
	assert.True(t, ret)

	close(quit)

	safeReceive(done)

	assert.Len(t, backend.packets, 2)
	assert.Equal(t, packet.SUBSCRIBE, backend.packets[0].Type())
	assert.Equal(t, packet.DISCONNECT, backend.packets[1].Type())
}
