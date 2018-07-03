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
func OfflineSubscriptionTest(t *testing.T, config *Config, id, topic string, qos uint8) {
	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearSession(options, 10*time.Second))

	offlineSubscriber := client.New()

	connectFuture, err := offlineSubscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := offlineSubscriber.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes())

	err = offlineSubscriber.Disconnect()
	assert.NoError(t, err)

	publisher := client.New()

	connectFuture, err = publisher.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	publishFuture, err := publisher.Publish(topic, testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	err = publisher.Disconnect()
	assert.NoError(t, err)

	wait := make(chan struct{})

	offlineReceiver := client.New()
	offlineReceiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err = offlineReceiver.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.True(t, connectFuture.SessionPresent())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = offlineReceiver.Disconnect()
	assert.NoError(t, err)
}

// OfflineSubscriptionRetainedTest tests the broker for properly handling
// retained messages and offline subscriptions.
func OfflineSubscriptionRetainedTest(t *testing.T, config *Config, id, topic string, qos uint8) {
	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearRetainedMessage(options, topic, 10*time.Second))
	assert.NoError(t, client.ClearSession(options, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	offlineSubscriber := client.New()

	connectFuture, err := offlineSubscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := offlineSubscriber.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes())

	err = offlineSubscriber.Disconnect()
	assert.NoError(t, err)

	publisher := client.New()

	connectFuture, err = publisher.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	publishFuture, err := publisher.Publish(topic, testPayload, qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	err = publisher.Disconnect()
	assert.NoError(t, err)

	wait := make(chan struct{})
	count := 0
	last := false

	offlineReceiver := client.New()
	offlineReceiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)

		if count == 0 {
			last = msg.Retain
		} else {
			assert.NotEqual(t, last, msg.Retain)
		}

		count++
		if count == 2 {
			close(wait)
		}

		return nil
	}

	connectFuture, err = offlineReceiver.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.True(t, connectFuture.SessionPresent())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = offlineReceiver.Disconnect()
	assert.NoError(t, err)
}
