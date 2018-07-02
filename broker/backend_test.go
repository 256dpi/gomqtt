package broker

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/spec"

	"github.com/stretchr/testify/assert"
)

func TestBrokerWithMemoryBackend(t *testing.T) {
	backend := NewMemoryBackend()
	backend.Credentials = map[string]string{
		"allow": "allow",
	}

	port, quit, done := Run(NewEngine(backend), "tcp")

	config := spec.AllFeatures()
	config.URL = "tcp://allow:allow@localhost:" + port
	config.DenyURL = "tcp://deny:deny@localhost:" + port
	config.ProcessWait = 20 * time.Millisecond
	config.NoMessageWait = 50 * time.Millisecond
	config.MessageRetainWait = 50 * time.Millisecond

	spec.Run(t, config)

	close(quit)

	safeReceive(done)
}

func TestMemoryBackendClose(t *testing.T) {
	backend := NewMemoryBackend()

	port, quit, done := Run(NewEngine(backend), "tcp")

	options := client.NewConfigWithClientID("tcp://localhost:"+port, "close1")
	options.CleanSession = false

	wait1 := make(chan struct{})
	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) error {
		assert.Nil(t, msg)
		assert.Error(t, err)
		close(wait1)

		return nil
	}

	connectFuture, err := client1.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	options.ClientID = "close2"
	options.CleanSession = true

	wait2 := make(chan struct{})
	client2 := client.New()
	client2.Callback = func(msg *packet.Message, err error) error {
		assert.Nil(t, msg)
		assert.Error(t, err)
		close(wait2)

		return nil
	}

	connectFuture, err = client2.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	ret := backend.Close(5 * time.Second)
	assert.True(t, ret)

	safeReceive(wait1)
	safeReceive(wait2)

	close(quit)

	safeReceive(done)
}
