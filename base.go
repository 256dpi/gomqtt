package spec

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func PublishSubscribeTest(t *testing.T, config *Config, out, in string, sub, pub uint8) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(out, testPayload, pub, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}

func UnsubscribeTest(t *testing.T, config *Config, topic string, qos uint8) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/2", msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, qos, msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(topic+"/1", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = client.Subscribe(topic+"/2", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client.Unsubscribe(topic + "/1")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	publishFuture, err := client.Publish(topic+"/1", testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	publishFuture, err = client.Publish(topic+"/2", testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}

func SubscriptionUpgradeTest(t *testing.T, config *Config, topic string, from, to uint8) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(to), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(topic, from)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{from}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = client.Subscribe(topic, to)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{to}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, to, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)
}

func OverlappingSubscriptionsTest(t *testing.T, config *Config, pub, sub string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, pub, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, byte(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(sub, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = client.Subscribe(pub, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(pub, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}

func MultipleSubscriptionTest(t *testing.T, config *Config, topic string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/3", msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: topic + "/1", QOS: 0},
		{Topic: topic + "/2", QOS: 1},
		{Topic: topic + "/3", QOS: 2},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0, 1, 2}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic+"/3", testPayload, 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}

func DuplicateSubscriptionTest(t *testing.T, config *Config, topic string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(1), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: topic, QOS: 0},
		{Topic: topic, QOS: 1},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0, 1}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, 1, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}

func IsolatedSubscriptionTest(t *testing.T, config *Config, topic string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic+"/foo", msg.Topic)
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

	subscribeFuture, err := client.Subscribe(topic+"/foo", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	publishFuture, err = client.Publish(topic+"/bar", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	publishFuture, err = client.Publish(topic+"/baz", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	publishFuture, err = client.Publish(topic+"/foo", testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = client.Disconnect()
	assert.NoError(t, err)
}

func WillTest(t *testing.T, config *Config, topic string, sub, pub uint8) {
	clientWithWill := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     pub,
	}

	connectFuture, err := clientWithWill.Connect(config.URL, opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	clientReceivingWill := client.New()
	wait := make(chan struct{})

	clientReceivingWill.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err = clientReceivingWill.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := clientReceivingWill.Subscribe(topic, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	err = clientWithWill.Close()
	assert.NoError(t, err)

	<-wait

	time.Sleep(config.NoMessageWait)

	err = clientReceivingWill.Disconnect()
	assert.NoError(t, err)
}

func CleanWillTest(t *testing.T, config *Config, topic string) {
	clientWithWill := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     0,
	}

	connectFuture, err := clientWithWill.Connect(config.URL, opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	nonReceiver := client.New()

	nonReceiver.Callback = func(msg *packet.Message, err error) {
		assert.Fail(t, "should not be called")
	}

	connectFuture, err = nonReceiver.Connect(config.URL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := nonReceiver.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	err = clientWithWill.Disconnect()
	assert.NoError(t, err)

	time.Sleep(config.NoMessageWait)

	err = nonReceiver.Disconnect()
	assert.NoError(t, err)
}

func KeepAliveTest(t *testing.T, config *Config) {
	opts := client.NewOptions()
	opts.KeepAlive = "2s" // mosquitto fails with a keep alive of 1s

	client := client.New()

	var reqCounter int32
	var respCounter int32

	client.Logger = func(message string) {
		if strings.Contains(message, "Pingreq") {
			atomic.AddInt32(&reqCounter, 1)
		} else if strings.Contains(message, "Pingresp") {
			atomic.AddInt32(&respCounter, 1)
		}
	}

	connectFuture, err := client.Connect(config.URL, opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	time.Sleep(4500 * time.Millisecond)

	err = client.Disconnect()
	assert.NoError(t, err)

	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCounter))
	assert.Equal(t, int32(2), atomic.LoadInt32(&respCounter))
}

func KeepAliveTimeoutTest(t *testing.T, config *Config) {
	username, password := config.usernamePassword()

	connect := packet.NewConnectPacket()
	connect.KeepAlive = 1
	connect.Username = username
	connect.Password = password

	connack := packet.NewConnackPacket()

	client := tools.NewFlow().
		Send(connect).
		Receive(connack).
		End()

	conn, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	client.Test(t, conn)
}
