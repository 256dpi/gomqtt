package spec

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

// RetainedMessageTest tests the broker for properly handling retained messages.
func RetainedMessageTest(t *testing.T, config *Config, out, in string, sub, pub packet.QOS) {
	assert.NoError(t, client.ClearRetainedMessage(client.NewConfig(config.URL), out, 10*time.Second))

	time.Sleep(config.MessageRetainWait)

	retainer := client.New()

	cf, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	pf, err := retainer.Publish(out, testPayload, pub, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()

	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, sub, msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	cf, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := receiver.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{sub}, sf.ReturnCodes())

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

	cf, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	pf, err := retainer.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	pf, err = retainer.Publish(topic, testPayload2, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()

	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload2, msg.Payload)
		assert.Equal(t, packet.QOS(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	cf, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

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

	cf, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	pf, err := retainer.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

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
		assert.Equal(t, packet.QOS(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	cf, err = receiverAndClearer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := receiverAndClearer.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	pf, err = receiverAndClearer.Publish(topic, nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = receiverAndClearer.Disconnect()
	assert.NoError(t, err)

	nonReceiver := client.New()
	nonReceiver.Callback = func(msg *packet.Message, err error) error {
		assert.Fail(t, "should not be called")
		return nil
	}

	cf, err = nonReceiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err = nonReceiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

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

	sf, err := c.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	time.Sleep(config.MessageRetainWait)

	pf, err := c.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

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

	sf, err := c.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	time.Sleep(config.MessageRetainWait)

	pf, err := c.Publish(topic, nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

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

	cf, err := clientWithRetainedWill.Connect(opts)
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	err = clientWithRetainedWill.Close()
	assert.NoError(t, err)

	time.Sleep(config.MessageRetainWait)

	receiver := client.New()
	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, packet.QOS(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
		return nil
	}

	cf, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

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

	cf, err := retainer.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	pf, err := retainer.Publish(topic, testPayload, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, pf.Wait(10*time.Second))

	time.Sleep(config.MessageRetainWait)

	err = retainer.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()
	wait := make(chan struct{}, 1)

	receiver.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, packet.QOS(0), msg.QOS)
		assert.True(t, msg.Retain)

		wait <- struct{}{}
		return nil
	}

	cf, err = receiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, cf.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, cf.ReturnCode())
	assert.False(t, cf.SessionPresent())

	sf, err := receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	safeReceive(wait)

	sf, err = receiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, sf.Wait(10*time.Second))
	assert.Equal(t, []packet.QOS{0}, sf.ReturnCodes())

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = receiver.Disconnect()
	assert.NoError(t, err)
}
