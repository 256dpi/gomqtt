package spec

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

// AuthenticationTest tests the broker for valid and invalid authentication.
func AuthenticationTest(t *testing.T, config *Config) {
	deniedClient := client.New()
	deniedClient.Callback = func(msg *packet.Message, err error) error {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
		return nil
	}

	cf, err := deniedClient.Connect(client.NewConfig(config.DenyURL))
	assert.NoError(t, err)
	assert.Error(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.NotAuthorized, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	allowedClient := client.New()

	cf, err = allowedClient.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	err = allowedClient.Disconnect()
	assert.NoError(t, err)
}

// UniqueClientIDUncleanTest tests the broker for enforcing unique client ids.
func UniqueClientIDUncleanTest(t *testing.T, config *Config) {
	id := config.clientID()

	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearSession(options, 10*time.Second))

	wait := make(chan struct{})

	firstClient := client.New()
	firstClient.Callback = func(msg *packet.Message, err error) error {
		close(wait)
		return nil
	}

	cf, err := firstClient.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	secondClient := client.New()

	cf, err = secondClient.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.True(t, cf.SessionPresent())

	safeReceive(wait)

	err = secondClient.Disconnect()
	assert.NoError(t, err)
}

// UniqueClientIDCleanTest tests the broker for enforcing unique client ids.
func UniqueClientIDCleanTest(t *testing.T, config *Config) {
	id := config.clientID()

	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = true

	wait := make(chan struct{})

	firstClient := client.New()
	firstClient.Callback = func(msg *packet.Message, err error) error {
		close(wait)
		return nil
	}

	cf, err := firstClient.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	secondClient := client.New()

	cf, err = secondClient.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	safeReceive(wait)

	err = secondClient.Disconnect()
	assert.NoError(t, err)
}

// RootSlashDistinctionTest tests the broker for supporting the root slash
// distinction.
func RootSlashDistinctionTest(t *testing.T, config *Config, topic string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, packet.QOS(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	cf, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := c.Subscribe("/"+topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	sf, err = c.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	pf, err := c.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}
