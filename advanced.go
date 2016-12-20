package spec

import (
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

// AuthenticationTest tests the broker for valid and invalid authentication.
func AuthenticationTest(t *testing.T, config *Config) {
	deniedClient := client.New()
	deniedClient.Callback = func(msg *packet.Message, err error) {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture, err := deniedClient.Connect(client.NewOptions(config.DenyURL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ErrNotAuthorized, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	allowedClient := client.New()

	connectFuture, err = allowedClient.Connect(client.NewOptions(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	err = allowedClient.Disconnect()
	assert.NoError(t, err)
}

// UniqueClientIDTest tests the broker for enforcing unique client ids.
func UniqueClientIDTest(t *testing.T, config *Config, id string) {
	options := client.NewOptionsWithClientID(config.URL, id)

	assert.NoError(t, client.ClearSession(options))

	wait := make(chan struct{})

	firstClient := client.New()
	firstClient.Callback = func(msg *packet.Message, err error) {
		assert.Error(t, err)
		close(wait)
	}

	connectFuture, err := firstClient.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	secondClient := client.New()

	connectFuture, err = secondClient.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	<-wait

	err = secondClient.Disconnect()
	assert.NoError(t, err)
}

// RootSlashDistinctionTest tests the broker for supporting the root slash
// distinction.
func RootSlashDistinctionTest(t *testing.T, config *Config, topic string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := c.Connect(client.NewOptions(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := c.Subscribe("/"+topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = c.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := c.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}
