package spec

import (
	"testing"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func AuthenticationTest(t *testing.T, config *Config, denyURL string) {
	deniedClient := client.New()
	deniedClient.Callback = func(msg *packet.Message, err error) {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture, err := deniedClient.Connect(denyURL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ErrNotAuthorized, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	allowedClient := client.New()

	connectFuture, err = allowedClient.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	err = allowedClient.Disconnect()
	assert.NoError(t, err)
}

func UniqueClientIDTest(t *testing.T, config *Config, id string) {
	assert.NoError(t, client.ClearSession(config.URL, id))

	options := client.NewOptions()
	options.ClientID = id

	wait := make(chan struct{})

	firstClient := client.New()
	firstClient.Callback = func(msg *packet.Message, err error) {
		assert.Error(t, err)
		close(wait)
	}

	connectFuture, err := firstClient.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	secondClient := client.New()

	connectFuture, err = secondClient.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	<-wait

	err = secondClient.Disconnect()
	assert.NoError(t, err)
}
