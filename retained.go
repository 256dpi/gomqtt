package spec

import (
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func RetainedMessageTest(t *testing.T, config *Config, out, in string, sub, pub uint8) {
	assert.NoError(t, client.ClearRetainedMessage(config.URL, out))

	client1 := client.New()

	connectFuture1, err := client1.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	publishFuture, err := client1.Publish(out, testPayload, pub, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	time.Sleep(config.MessageRetainWait)

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)
}

func ClearRetainedMessageTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(config.URL, topic))

	// client1 retains message

	client1 := client.New()

	connectFuture1, err := client1.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	publishFuture1, err := client1.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture1.Wait())

	time.Sleep(config.MessageRetainWait)

	err = client1.Disconnect()
	assert.NoError(t, err)

	// client2 receives retained message and clears it

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		// ignore directly send message
		if msg.Topic == topic && msg.Payload == nil {
			return
		}

		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture1, err := client2.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture1.ReturnCodes)

	<-wait

	publishFuture2, err := client2.Publish(topic, nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	time.Sleep(config.MessageRetainWait)

	err = client2.Disconnect()
	assert.NoError(t, err)

	// client3 should not receive any message

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		assert.Fail(t, "should not be called")
	}

	connectFuture3, err := client3.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.False(t, connectFuture3.SessionPresent)

	subscribeFuture2, err := client3.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture2.ReturnCodes)

	time.Sleep(config.NoMessageWait)

	err = client3.Disconnect()
	assert.NoError(t, err)
}

func DirectRetainedMessageTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(config.URL, topic))

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

	subscribeFuture, err := client.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)
}

func RetainedWillTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(config.URL, topic))

	// client1 connects with a retained will and dies

	client1 := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     0,
		Retain:  true,
	}

	connectFuture1, err := client1.Connect(config.URL, opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	err = client1.Close()
	assert.NoError(t, err)

	time.Sleep(config.MessageRetainWait)

	// client2 subscribes to the wills topic and receives the retained will

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)
}
