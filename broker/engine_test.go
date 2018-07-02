package broker

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"

	"github.com/stretchr/testify/assert"
)

func TestConnectTimeout(t *testing.T) {
	engine := NewEngine(NewMemoryBackend())
	engine.ConnectTimeout = 10 * time.Millisecond

	port, quit, done := Run(engine, "tcp")

	conn, err := transport.Dial("tcp://localhost:" + port)
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	close(quit)
	safeReceive(done)
}

func TestDefaultReadLimit(t *testing.T) {
	engine := NewEngine(NewMemoryBackend())
	engine.DefaultReadLimit = 1

	port, quit, done := Run(engine, "tcp")

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.Error(t, err)
		close(wait)
		return nil
	}

	cf, err := c.Connect(client.NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.Error(t, cf.Wait(10*time.Second))

	safeReceive(wait)
	close(quit)
	safeReceive(done)
}
