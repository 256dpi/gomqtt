package spec

import (
	"testing"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func OfflineSubscriptionTest(t *testing.T, config *Config, id, topic string, qos uint8) {
	assert.NoError(t, client.ClearSession(config.URL, id))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	offlineSubscribe := client.New()

	connectFuture, err := offlineSubscribe.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := offlineSubscribe.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = offlineSubscribe.Disconnect()
	assert.NoError(t, err)

	publisher := client.New()

	connectFuture, err = publisher.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	publishFuture, err := publisher.Publish(topic, testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = publisher.Disconnect()
	assert.NoError(t, err)

	wait := make(chan struct{})

	offlineReceiver := client.New()
	offlineReceiver.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err = offlineReceiver.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.True(t, connectFuture.SessionPresent)

	<-wait

	err = offlineReceiver.Disconnect()
	assert.NoError(t, err)
}

func OfflineSubscriptionRetainedTest(t *testing.T, config *Config, id, topic string, qos uint8) {
	assert.NoError(t, client.ClearSession(config.URL, id))
	assert.NoError(t, client.ClearRetainedMessage(config.URL, topic))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	offlineSubscriber := client.New()

	connectFuture, err := offlineSubscriber.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := offlineSubscriber.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = offlineSubscriber.Disconnect()
	assert.NoError(t, err)

	publisher := client.New()

	connectFuture, err = publisher.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	publishFuture, err := publisher.Publish(topic, testPayload, qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = publisher.Disconnect()
	assert.NoError(t, err)

	wait := make(chan struct{})

	offlineReceiver := client.New()
	offlineReceiver.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err = offlineReceiver.Connect(config.URL, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.True(t, connectFuture.SessionPresent)

	<-wait

	err = offlineReceiver.Disconnect()
	assert.NoError(t, err)
}
