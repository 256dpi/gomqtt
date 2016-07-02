package spec

import (
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func AuthenticationTest(t *testing.T, config *Config) {
	deniedClient := client.New()
	deniedClient.Callback = func(msg *packet.Message, err error) {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture, err := deniedClient.Connect(config.DenyURL, nil)
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

func RootSlashDistinctionTest(t *testing.T, config *Config, topic string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe("/"+topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = client.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}
