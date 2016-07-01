package spec

import (
	"testing"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func OfflineSubscriptionTest(t *testing.T, url string, id, topic string, qos uint8) {
	assert.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	/* offline subscriber */

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	/* publisher */

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish(topic, testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(url, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	assert.NoError(t, err)
}

func OfflineSubscriptionRetainedTest(t *testing.T, url string, id, topic string, qos uint8) {
	assert.NoError(t, client.ClearSession(url, id))
	assert.NoError(t, client.ClearRetainedMessage(url, topic))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	/* offline subscriber */

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	/* publisher */

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish(topic, testPayload, qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(url, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	assert.NoError(t, err)
}
