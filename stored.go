package spec

import (
	"testing"
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

// PublishResendQOS1Test tests the broker for properly retrying QOS1 publish
// packets.
func PublishResendQOS1Test(t *testing.T, config *Config, id, topic string) {
	assert.NoError(t, client.ClearSession(client.NewConfigWithClientID(config.URL, id)))

	username, password := config.usernamePassword()

	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = id
	connect.Username = username
	connect.Password = password

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: topic, QOS: 1},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = topic
	publishOut.Message.QOS = 1

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = topic
	publishIn.Message.QOS = 1

	pubackIn := packet.NewPubackPacket()
	pubackIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	conn1, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	err = tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Send(subscribe).
		Skip(). // suback
		Send(publishOut).
		Skip(). // puback
		Receive(publishIn).
		Close().
		Test(conn1)
	assert.NoError(t, err)

	conn2, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	err = tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(publishIn).
		Send(pubackIn).
		Send(disconnect).
		Close().
		Test(conn2)
	assert.NoError(t, err)
}

// PublishResendQOS2Test tests the broker for properly retrying QOS2 Publish
// packets.
func PublishResendQOS2Test(t *testing.T, config *Config, id, topic string) {
	assert.NoError(t, client.ClearSession(client.NewConfigWithClientID(config.URL, id)))

	username, password := config.usernamePassword()

	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = id
	connect.Username = username
	connect.Password = password

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: topic, QOS: 2},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = topic
	publishOut.Message.QOS = 2

	pubrelOut := packet.NewPubrelPacket()
	pubrelOut.PacketID = 2

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = topic
	publishIn.Message.QOS = 2

	pubrecIn := packet.NewPubrecPacket()
	pubrecIn.PacketID = 1

	pubcompIn := packet.NewPubcompPacket()
	pubcompIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	conn1, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	err = tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Send(subscribe).
		Skip(). // suback
		Send(publishOut).
		Skip(). // pubrec
		Send(pubrelOut).
		Skip(). // pubcomp
		Receive(publishIn).
		Close().
		Test(conn1)
	assert.NoError(t, err)

	time.Sleep(config.ProcessWait)

	conn2, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	err = tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(publishIn).
		Send(pubrecIn).
		Skip(). // pubrel
		Send(pubcompIn).
		Send(disconnect).
		Close().
		Test(conn2)
	assert.NoError(t, err)
}

// PubrelResendQOS2Test tests the broker for properly retrying QOS2 Pubrel
// packets.
func PubrelResendQOS2Test(t *testing.T, config *Config, id, topic string) {
	assert.NoError(t, client.ClearSession(client.NewConfigWithClientID(config.URL, id)))

	username, password := config.usernamePassword()

	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = id
	connect.Username = username
	connect.Password = password

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: topic, QOS: 2},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = topic
	publishOut.Message.QOS = 2

	pubrelOut := packet.NewPubrelPacket()
	pubrelOut.PacketID = 2

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = topic
	publishIn.Message.QOS = 2

	pubrecIn := packet.NewPubrecPacket()
	pubrecIn.PacketID = 1

	pubrelIn := packet.NewPubrelPacket()
	pubrelIn.PacketID = 1

	pubcompIn := packet.NewPubcompPacket()
	pubcompIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	conn1, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	err = tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Send(subscribe).
		Skip(). // suback
		Send(publishOut).
		Skip(). // pubrec
		Send(pubrelOut).
		Skip(). // pubcomp
		Receive(publishIn).
		Send(pubrecIn).
		Close().
		Test(conn1)
	assert.NoError(t, err)

	time.Sleep(config.ProcessWait)

	conn2, err := transport.Dial(config.URL)
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	err = tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(pubrelIn).
		Send(pubcompIn).
		Send(disconnect).
		Close().
		Test(conn2)
	assert.NoError(t, err)
}

// StoredSubscriptionsTest tests the broker for properly handling stored
// subscriptions.
func StoredSubscriptionsTest(t *testing.T, config *Config, id, topic string, qos uint8) {
	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearSession(options))

	subscriber := client.New()

	connectFuture, err := subscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := subscriber.Subscribe(topic, qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = subscriber.Disconnect()
	assert.NoError(t, err)

	receiver := client.New()

	wait := make(chan struct{})

	receiver.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, testPayload, msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err = receiver.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.True(t, connectFuture.SessionPresent)

	publishFuture, err := receiver.Publish(topic, testPayload, qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	time.Sleep(config.NoMessageWait)

	err = receiver.Disconnect()
	assert.NoError(t, err)
}

// CleanStoredSubscriptionsTest tests the broker for properly clearing stored
// subscriptions.
func CleanStoredSubscriptionsTest(t *testing.T, config *Config, id, topic string) {
	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearSession(options))

	subscriber := client.New()

	connectFuture, err := subscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := subscriber.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	err = subscriber.Disconnect()
	assert.NoError(t, err)

	nonReceiver := client.New()
	nonReceiver.Callback = func(msg *packet.Message, err error) {
		assert.Fail(t, "should not be called")
	}

	options.CleanSession = true

	connectFuture, err = nonReceiver.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	publishFuture, err := nonReceiver.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	time.Sleep(config.NoMessageWait)

	err = nonReceiver.Disconnect()
	assert.NoError(t, err)
}

// RemoveStoredSubscriptionTest tests the broker for properly removing stored
// subscriptions.
func RemoveStoredSubscriptionTest(t *testing.T, config *Config, id, topic string) {
	options := client.NewConfigWithClientID(config.URL, id)
	options.CleanSession = false

	assert.NoError(t, client.ClearSession(options))

	subscriberAndUnsubscriber := client.New()

	connectFuture, err := subscriberAndUnsubscriber.Connect(options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := subscriberAndUnsubscriber.Subscribe(topic, 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := subscriberAndUnsubscriber.Unsubscribe(topic)
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	err = subscriberAndUnsubscriber.Disconnect()
	assert.NoError(t, err)

	nonReceiver := client.New()
	nonReceiver.Callback = func(msg *packet.Message, err error) {
		assert.Fail(t, "should not be called")
	}

	connectFuture, err = nonReceiver.Connect(client.NewConfig(config.URL))
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	publishFuture, err := nonReceiver.Publish(topic, testPayload, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	time.Sleep(config.NoMessageWait)

	err = nonReceiver.Disconnect()
	assert.NoError(t, err)
}
