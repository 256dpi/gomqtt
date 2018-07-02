package spec

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

// RetainedMessageTest tests the broker for properly handling retained messages.
func RetainedMessageTest(t *testing.T, config *Config, out, in string, sub, pub uint8) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), out, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	retainer := client.New()

	connectFuture, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	publishFuture, err := retainer.Publish(out, testPayload, pub, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()

	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := receiver.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = receiver.Disconnect()
	assert.NoError(t, err)
}

// RetainedMessageReplaceTest tests the broker for replacing existing retained
// messages.
func RetainedMessageReplaceTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), topic, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	retainer := client.New()

	connectFuture, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	publishFuture, err := retainer.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	publishFuture, err = retainer.Publish(topic, testPayload2, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()

	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload2, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = receiver.Disconnect()
	assert.NoError(t, err)
}

// ClearRetainedMessageTest tests the broker for clearing retained messages.
func ClearRetainedMessageTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), topic, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	retainer := client.New()

	connectFuture, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	publishFuture, err := retainer.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiverAndClearer := client.New()

	wait := make(chan struct{})

	receiverAndClearer.Callback = func(msg *packet.Message, err error) error {
		// ignore directly send message
		if msg.Topic == topic && msg.Payload == nil {
			return nil
		}

		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err = receiverAndClearer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := receiverAndClearer.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	publishFuture, err = receiverAndClearer.Publish(topic, nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = receiverAndClearer.Disconnect()
	assert.NoError(t, err)

	nonReceiver := client.New()
	nonReceiver.Callback = func(msg *packet.Message, err error) error {
		assert.Fail(t, "should not be called")
		return nil
	}

	connectFuture, err = nonReceiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err = nonReceiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	time.Sleep(config.NoMessageWait)

	err = nonReceiver.Disconnect()
	assert.NoError(t, err)
}

// DirectRetainedMessageTest tests the broker for properly handling subscriptions
// with retained messages.
func DirectRetainedMessageTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), topic, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := c.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// DirectClearRetainedMessageTest tests the broker for properly dispatching a
// messages intended to clear a retained message.
func DirectClearRetainedMessageTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), topic, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Nil(t, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := c.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(topic, nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// RetainedWillTest tests the broker for support of retained will messages.
func RetainedWillTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), topic, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	clientWithRetainedWill := client.New()

	opts := client.NewConfig(config.URL)
	opts.WillMessage = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     0,
		Retain:  true,
	}

	connectFuture, err := clientWithRetainedWill.Connect(opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	err = clientWithRetainedWill.Close()
	assert.NoError(t, err)

	time.Sleep(config.MessageRetainWait)

	receiver := client.New()
	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = receiver.Disconnect()
	assert.NoError(t, err)
}

// RetainedMessageResubscriptionTest tests the broker for properly dispatching
// retained messages on resubscription.
func RetainedMessageResubscriptionTest(t *testing.T, config *Config, topic string) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), topic, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	retainer := client.New()

	connectFuture, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	publishFuture, err := retainer.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()

	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		wait <- struct{}{}
		return nil
	}

	connectFuture, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	safeReceive(wait)

	subscribeFuture, err = receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = receiver.Disconnect()
	assert.NoError(t, err)
}
