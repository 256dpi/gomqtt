package spec

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

// OfflineSubscriptionTest tests the broker for properly handling offline
// subscriptions.
func OfflineSubscriptionTest(t *testing.T, config *Config, topic string, sub, pub packet.QOS, await bool) {
	id := config.clientID()

	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearSession(options, 10*time.Second))

	offlineSubscriber := client.New()

	cf, err := offlineSubscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := offlineSubscriber.Subscribe(topic, sub)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{sub}, sf.ReturnCodes())

	err = offlineSubscriber.Disconnect()
	assert.NoError(t, err)

	publisher := client.New()

	cf, err = publisher.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	pf, err := publisher.Publish(topic, testPayload, pub, false)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	err = publisher.Disconnect()
	assert.NoError(t, err)

	wait := make(chan struct{})

	offlineReceiver := client.New()
	offlineReceiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, lower(sub, pub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	cf, err = offlineReceiver.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.True(t, cf.SessionPresent())

	if await {
		safeReceive(wait)
	}

	time.Sleep(config.NoMessageWait)

	err = offlineReceiver.Disconnect()
	assert.NoError(t, err)
}

// OfflineSubscriptionRetainedTest tests the broker for properly handling
// retained messages and offline subscriptions.
func OfflineSubscriptionRetainedTest(t *testing.T, config *Config, topic string, sub, pub packet.QOS, await bool) {
	id := config.clientID()

	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearRetainedMessage(options, topic, 10*time.Second))
	assert.NoError(t, client.ClearSession(options, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	offlineSubscriber := client.New()

	cf, err := offlineSubscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := offlineSubscriber.Subscribe(topic, sub)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{sub}, sf.ReturnCodes())

	err = offlineSubscriber.Disconnect()
	assert.NoError(t, err)

	publisher := client.New()

	cf, err = publisher.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	pf, err := publisher.Publish(topic, testPayload, pub, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = publisher.Disconnect()
	assert.NoError(t, err)

	wait := make(chan struct{})

	offlineReceiver := client.New()
	offlineReceiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, sub, msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)

		return nil
	}

	cf, err = offlineReceiver.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.True(t, cf.SessionPresent())

	if await {
		safeReceive(wait)
	}

	time.Sleep(config.NoMessageWait)

	err = offlineReceiver.Disconnect()
	assert.NoError(t, err)
}
