package spec

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/256dpi/gomqtt/transport/flow"
	"github.com/stretchr/testify/assert"
)

// PublishSubscribeTest tests the broker for basic pub sub support.
func PublishSubscribeTest(t *testing.T, config *Config, pub, sub string, subQOS, pubQOS uint8) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, pub, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(subQOS), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := c.Subscribe(sub, subQOS)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{subQOS}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(pub, testPayload, pubQOS, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// UnsubscribeTest tests the broker for unsubscribe support.
func UnsubscribeTest(t *testing.T, config *Config, topic string, qos uint8) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/2", msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, qos, msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := c.Subscribe(topic+"/1", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes())

	subscribeFuture, err = c.Subscribe(topic+"/2", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes())

	unsubscribeFuture, err := c.Unsubscribe(topic + "/1")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait(10*time.Second))

	publishFuture, err := c.Publish(topic+"/1", testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	publishFuture, err = c.Publish(topic+"/2", testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// UnsubscribeNotExistingSubscriptionTest tests the broker for allowing
// unsubscribing not existing topics.
func UnsubscribeNotExistingSubscriptionTest(t *testing.T, config *Config, topic string) {
	c := client.New()

	c.Callback = func(msg *packet.Message, err error) error {
		assert.Fail(t, "should not be called")
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	unsubscribeFuture, err := c.Unsubscribe(topic)
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait(10*time.Second))

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// UnsubscribeOverlappingSubscriptions tests the broker for properly unsubscribing
// overlapping topics.
func UnsubscribeOverlappingSubscriptions(t *testing.T, config *Config, topic string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/foo", msg.Topic)
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

	subscribeFuture, err := c.Subscribe(topic+"/#", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	subscribeFuture, err = c.Subscribe(topic+"/+", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	unsubscribeFuture, err := c.Unsubscribe(topic + "/#")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait(10*time.Second))

	publishFuture, err := c.Publish(topic+"/foo", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// SubscriptionUpgradeTest tests the broker for properly upgrading subscriptions,
func SubscriptionUpgradeTest(t *testing.T, config *Config, topic string, from, to uint8) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(to), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := c.Subscribe(topic, from)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{from}, subscribeFuture.ReturnCodes())

	subscribeFuture, err = c.Subscribe(topic, to)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{to}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(topic, testPayload, to, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// OverlappingSubscriptionsTest tests the broker for properly handling overlapping
// subscriptions.
func OverlappingSubscriptionsTest(t *testing.T, config *Config, pub, sub string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, pub, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, byte(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := c.Subscribe(sub, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	subscribeFuture, err = c.Subscribe(pub, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(pub, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// MultipleSubscriptionTest tests the broker for properly handling multiple
// subscriptions.
func MultipleSubscriptionTest(t *testing.T, config *Config, topic string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/3", msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subs := []packet.Subscription{
		{Topic: topic + "/1", QOS: 0},
		{Topic: topic + "/2", QOS: 1},
		{Topic: topic + "/3", QOS: 2},
	}

	subscribeFuture, err := c.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0, 1, 2}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(topic+"/3", testPayload, 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// DuplicateSubscriptionTest tests the broker for properly handling duplicate
// subscriptions.
func DuplicateSubscriptionTest(t *testing.T, config *Config, topic string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(1), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err := c.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subs := []packet.Subscription{
		{Topic: topic, QOS: 0},
		{Topic: topic, QOS: 1},
	}

	subscribeFuture, err := c.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0, 1}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(topic, testPayload, 1, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// IsolatedSubscriptionTest tests the broker for properly isolating subscriptions.
func IsolatedSubscriptionTest(t *testing.T, config *Config, topic string) {
	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/foo", msg.Topic)
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

	subscribeFuture, err := c.Subscribe(topic+"/foo", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	publishFuture, err := c.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	publishFuture, err = c.Publish(topic+"/bar", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	publishFuture, err = c.Publish(topic+"/baz", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	publishFuture, err = c.Publish(topic+"/foo", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait(10*time.Second))

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = c.Disconnect()
	assert.NoError(t, err)
}

// WillTest tests the broker for supporting will messages.
func WillTest(t *testing.T, config *Config, topic string, sub, pub uint8) {
	clientWithWill := client.New()

	opts := client.NewConfig(config.URL)
	opts.WillMessage = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     pub,
	}

	connectFuture, err := clientWithWill.Connect(opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	clientReceivingWill := client.New()
	wait := make(chan struct{})

	clientReceivingWill.Callback = func(msg *packet.Message, err error) error {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
		return nil
	}

	connectFuture, err = clientReceivingWill.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	subscribeFuture, err := clientReceivingWill.Subscribe(topic, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes())

	err = clientWithWill.Close()
	assert.NoError(t, err)

	safeReceive(wait)

	time.Sleep(config.NoMessageWait)

	err = clientReceivingWill.Disconnect()
	assert.NoError(t, err)
}

// CleanWillTest tests the broker for properly handling will messages on a clean
// disconnect.
func CleanWillTest(t *testing.T, config *Config, topic string) {
	clientWithWill := client.New()

	opts := client.NewConfig(config.URL)
	opts.WillMessage = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     0,
	}

	connectFuture, err := clientWithWill.Connect(opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

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

	subscribeFuture, err := nonReceiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait(10*time.Second))
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes())

	err = clientWithWill.Disconnect()
	assert.NoError(t, err)

	time.Sleep(config.NoMessageWait)

	err = nonReceiver.Disconnect()
	assert.NoError(t, err)
}

// KeepAliveTest tests the broker for proper keep alive support.
func KeepAliveTest(t *testing.T, config *Config) {
	opts := client.NewConfig(config.URL)
	opts.KeepAlive = "2s" // mosquitto fails with a keep alive of 1s

	c := client.New()

	var reqCounter int32
	var respCounter int32

	c.Logger = func(message string) {
		if strings.Contains(message, "Pingreq") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "Pingresp") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	connectFuture, err := c.Connect(opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait(10*time.Second))
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode())
	assert.False(t, connectFuture.SessionPresent())

	time.Sleep(4500 * time.Millisecond)

	err = c.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(2), atomic.LoadInt32(&respCounter))
}

// KeepAliveTimeoutTest tests the broker for proper keep alive timeout detection
// support.
func KeepAliveTimeoutTest(t *testing.T, config *Config) {
	username, password := config.usernamePassword()

	connect := packet.NewConnectPacket()
	connect.KeepAlive = 1
	connect.Username = username
	connect.Password = password

	connack := packet.NewConnackPacket()

	c := flow.New().
		Send(connect).
		Receive(connack).
		End()

	conn, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	err = c.Test(conn)
	assert.NoError(t, err)
}
