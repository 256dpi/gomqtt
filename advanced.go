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

	connectFuture1, err := deniedClient.Connect(denyURL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ErrNotAuthorized, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	allowedClient := client.New()

	connectFuture2, err := allowedClient.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

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

	connectFuture1, err := firstClient.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	secondClient := client.New()

	connectFuture2, err := secondClient.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	<-wait

	err = secondClient.Disconnect()
	assert.NoError(t, err)
}
