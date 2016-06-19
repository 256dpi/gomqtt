package broker

import (
	"fmt"
	"testing"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

type SpecMatrix struct {
	RetainedMessages bool
	StoredSubscriptions bool
	OfflineSubscriptions bool
	UniqueClientIDs bool
}

var FullSpecMatrix = SpecMatrix{
	RetainedMessages: true,
	StoredSubscriptions: true,
	OfflineSubscriptions: true,
	UniqueClientIDs: true,
}

// Spec will fully test a Broker with its Backend and Session implementation to
// support all specified features in the matrix. The passed builder callback
// should always return a fresh instances of the Broker which should only allow the
// "allow:allow" login.
func Spec(t *testing.T, matrix SpecMatrix, builder func() *Broker) {
	println("Running Broker Publish Subscribe Test (QOS 0)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 0, 0)

	println("Running Broker Publish Subscribe Test (QOS 1)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 1, 1)

	println("Running Broker Publish Subscribe Test (QOS 2)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 2, 2)

	println("Running Broker Publish Subscribe Test (Wildcard One)")
	brokerPublishSubscribeTest(t, builder(), "foo/bar", "foo/+", 0, 0)

	println("Running Broker Publish Subscribe Test (Wildcard Some)")
	brokerPublishSubscribeTest(t, builder(), "foo/bar", "#", 0, 0)

	println("Running Broker Publish Subscribe Test (QOS Downgrade 1->0)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 0, 1)

	println("Running Broker Publish Subscribe Test (QOS Downgrade 2->0)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 0, 2)

	println("Running Broker Publish Subscribe Test (QOS Downgrade 2->1)")
	brokerPublishSubscribeTest(t, builder(), "test", "test", 1, 2)

	println("Running Broker Unsubscribe Test (QOS 0)")
	brokerUnsubscribeTest(t, builder(), 0)

	println("Running Broker Unsubscribe Test (QOS 1)")
	brokerUnsubscribeTest(t, builder(), 1)

	println("Running Broker Unsubscribe Test (QOS 2)")
	brokerUnsubscribeTest(t, builder(), 2)

	println("Running Broker Subscription Upgrade Test (QOS 0->1)")
	brokerSubscriptionUpgradeTest(t, builder(), 0, 1)

	println("Running Broker Subscription Upgrade Test (QOS 1->2)")
	brokerSubscriptionUpgradeTest(t, builder(), 1, 2)

	println("Running Broker Will Test (QOS 0)")
	brokerWillTest(t, builder(), 0, 0)

	println("Running Broker Will Test (QOS 1)")
	brokerWillTest(t, builder(), 1, 1)

	println("Running Broker Will Test (QOS 2)")
	brokerWillTest(t, builder(), 2, 2)

	// TODO: Test Clean Disconnect without forwarding the will.

	println("Running Broker Authentication Test")
	brokerAuthenticationTest(t, builder())

	println("Running Broker Multiple Subscription Test")
	brokerMultipleSubscriptionTest(t, builder())

	println("Running Broker Duplicate Subscription Test")
	brokerDuplicateSubscriptionTest(t, builder())

	println("Running Broker Publish Resend Test (QOS 1)")
	brokerPublishResendTestQOS1(t, builder())

	println("Running Broker Publish Resend Test (QOS 2)")
	brokerPublishResendTestQOS2(t, builder())

	println("Running Broker Pubrel Resend Test (QOS 2)")
	brokerPubrelResendTestQOS2(t, builder())

	if matrix.RetainedMessages {
		println("Running Broker Retained Message Test (QOS 0)")
		brokerRetainedMessageTest(t, builder(), "test", "test", 0, 0)

		println("Running Broker Retained Message Test (QOS 1)")
		brokerRetainedMessageTest(t, builder(), "test", "test", 1, 1)

		println("Running Broker Retained Message Test (QOS 2)")
		brokerRetainedMessageTest(t, builder(), "test", "test", 2, 2)

		println("Running Broker Retained Message Test (Wildcard One)")
		brokerRetainedMessageTest(t, builder(), "foo/bar", "foo/+", 0, 0)

		println("Running Broker Retained Message Test (Wildcard Some)")
		brokerRetainedMessageTest(t, builder(), "foo/bar", "#", 0, 0)

		println("Running Broker Clear Retained Message Test")
		brokerClearRetainedMessageTest(t, builder())

		println("Running Broker Direct Retained Message Test")
		brokerDirectRetainedMessageTest(t, builder())

		println("Running Broker Retained Will Test)")
		brokerRetainedWillTest(t, builder())
	}

	if matrix.StoredSubscriptions {
		println("Running Broker Stored Subscriptions Test (QOS 0)")
		brokerStoredSubscriptionsTest(t, builder(), 0)

		println("Running Broker Stored Subscriptions Test (QOS 1)")
		brokerStoredSubscriptionsTest(t, builder(), 1)

		println("Running Broker Stored Subscriptions Test (QOS 2)")
		brokerStoredSubscriptionsTest(t, builder(), 2)

		println("Running Broker Clean Stored Subscriptions Test")
		brokerCleanStoredSubscriptions(t, builder())

		println("Running Broker Remove Stored Subscription Test")
		brokerRemoveStoredSubscription(t, builder())
	}

	if matrix.OfflineSubscriptions {
		println("Running Broker Offline Subscription Test (QOS 1)")
		brokerOfflineSubscriptionTest(t, builder(), 1)

		println("Running Broker Offline Subscription Test (QOS 2)")
		brokerOfflineSubscriptionTest(t, builder(), 2)

		println("Running Broker Offline Subscription Test Retained (QOS 1)")
		brokerOfflineSubscriptionRetainedTest(t, builder(), 1)

		println("Running Broker Offline Subscription Test Retained (QOS 2)")
		brokerOfflineSubscriptionRetainedTest(t, builder(), 2)
	}

	if matrix.UniqueClientIDs {
		println("Running Broker Unique Client ID Test")
		brokerUniqueClientIDTest(t, builder())
	}
}

// TODO: Delivers old Wills in case of a crash.
// TODO: Add Reboot Persistence Test?

func runBroker(t *testing.T, broker *Broker, num int) (*tools.Port, chan struct{}) {
	port := tools.NewPort()

	server, err := transport.Launch(port.URL())
	assert.NoError(t, err)

	done := make(chan struct{})

	go func() {
		for i := 0; i < num; i++ {
			conn, err := server.Accept()
			assert.NoError(t, err)

			broker.Handle(conn)
		}

		err := server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	return port, done
}

func permittedURL(port *tools.Port) string {
	return fmt.Sprintf("tcp://allow:allow@localhost:%s/", port.Port())
}

func brokerPublishSubscribeTest(t *testing.T, broker *Broker, out, in string, sub, pub uint8) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(in, sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(out, []byte("test"), pub, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerRetainedMessageTest(t *testing.T, broker *Broker, out, in string, sub, pub uint8) {
	port, done := runBroker(t, broker, 2)

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	publishFuture, err := client1.Publish(out, []byte("test"), pub, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, out, msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
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

	<-done
}

func brokerClearRetainedMessageTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 3)

	// client1 retains message

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	publishFuture1, err := client1.Publish("test", []byte("test1"), 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture1.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	// client2 receives retained message and clears it

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test1"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture1, err := client2.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture1.ReturnCodes)

	<-wait

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	// client3 should not receive any message

	client3 := client.New()

	// TODO: Test non-receivement?

	connectFuture3, err := client3.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.False(t, connectFuture3.SessionPresent)

	subscribeFuture2, err := client3.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture2.ReturnCodes)

	err = client3.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerDirectRetainedMessageTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish("test", []byte("test"), 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerWillTest(t *testing.T, broker *Broker, sub, pub uint8) {
	port, done := runBroker(t, broker, 2)

	// client1 connects with a will

	client1 := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     pub,
	}

	connectFuture1, err := client1.Connect(permittedURL(port), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	// client2 subscribe to the wills topic

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(sub), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe("test", sub)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	// client1 dies

	err = client1.Close()
	assert.NoError(t, err)

	// client2 should receive the message

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerRetainedWillTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 2)

	// client1 connects with a retained will and dies

	client1 := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		QOS:     0,
		Retain:  true,
	}

	connectFuture1, err := client1.Connect(permittedURL(port), opts)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	err = client1.Close()
	assert.NoError(t, err)

	// client2 subscribes to the wills topic and receives the retained will

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerUnsubscribeTest(t *testing.T, broker *Broker, qos uint8) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test2", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, qos, msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe("test1", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = client.Subscribe("test2", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client.Unsubscribe("test1")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	publishFuture, err := client.Publish("test1", []byte("test"), qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	publishFuture, err = client.Publish("test2", []byte("test"), qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerSubscriptionUpgradeTest(t *testing.T, broker *Broker, from, to uint8) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(to), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subscribeFuture1, err := client.Subscribe("test", from)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture1.Wait())
	assert.Equal(t, []uint8{from}, subscribeFuture1.ReturnCodes)

	subscribeFuture2, err := client.Subscribe("test", to)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture2.Wait())
	assert.Equal(t, []uint8{to}, subscribeFuture2.ReturnCodes)

	publishFuture, err := client.Publish("test", []byte("test"), to, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerAuthenticationTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 2)

	// client1 should be denied

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture1, err := client1.Connect(port.URL(), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ErrNotAuthorized, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	// client2 should be allowed

	client2 := client.New()

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerMultipleSubscriptionTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test3", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(2), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: "test1", QOS: 0},
		{Topic: "test2", QOS: 1},
		{Topic: "test3", QOS: 2},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0, 1, 2}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish("test3", []byte("test"), 2, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerDuplicateSubscriptionTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 1)

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(1), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	assert.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: "test", QOS: 0},
		{Topic: "test", QOS: 1},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0, 1}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish("test", []byte("test"), 1, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerStoredSubscriptionsTest(t *testing.T, broker *Broker, qos uint8) {
	port, done := runBroker(t, broker, 2)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.True(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish("test", []byte("test"), qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerCleanStoredSubscriptions(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 2)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	options.CleanSession = true

	client2 := client.New()

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	// TODO: Test non-receivement?

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerRemoveStoredSubscription(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 2)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", 0)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client1.Unsubscribe("test")
	assert.NoError(t, err)
	assert.NoError(t, unsubscribeFuture.Wait())

	err = client1.Disconnect()
	assert.NoError(t, err)

	client2 := client.New()

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture2, err := client2.Publish("test", nil, 0, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture2.Wait())

	// TODO: Test non-receivement?

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerPublishResendTestQOS1(t *testing.T, broker *Broker) {
	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = "test"
	connect.Username = "allow"
	connect.Password = "allow"

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test", QOS: 1},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = "test"
	publishOut.Message.QOS = 1

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = "test"
	publishIn.Message.QOS = 1

	pubackIn := packet.NewPubackPacket()
	pubackIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	port, done := runBroker(t, broker, 2)

	conn1, err := transport.Dial(permittedURL(port))
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Send(subscribe).
		Skip(). // suback
		Send(publishOut).
		Skip(). // puback
		Receive(publishIn).
		Close().
		Test(t, conn1)

	conn2, err := transport.Dial(permittedURL(port))
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(publishIn).
		Send(pubackIn).
		Send(disconnect).
		Close().
		Test(t, conn2)

	<-done
}

func brokerPublishResendTestQOS2(t *testing.T, broker *Broker) {
	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = "test"
	connect.Username = "allow"
	connect.Password = "allow"

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test", QOS: 2},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = "test"
	publishOut.Message.QOS = 2

	pubrelOut := packet.NewPubrelPacket()
	pubrelOut.PacketID = 2

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = "test"
	publishIn.Message.QOS = 2

	pubrecIn := packet.NewPubrecPacket()
	pubrecIn.PacketID = 1

	pubcompIn := packet.NewPubcompPacket()
	pubcompIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	port, done := runBroker(t, broker, 2)

	conn1, err := transport.Dial(permittedURL(port))
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	tools.NewFlow().
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
		Test(t, conn1)

	conn2, err := transport.Dial(permittedURL(port))
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(publishIn).
		Send(pubrecIn).
		Skip(). // pubrel
		Send(pubcompIn).
		Send(disconnect).
		Close().
		Test(t, conn2)

	<-done
}

func brokerPubrelResendTestQOS2(t *testing.T, broker *Broker) {
	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = "test"
	connect.Username = "allow"
	connect.Password = "allow"

	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test", QOS: 2},
	}

	publishOut := packet.NewPublishPacket()
	publishOut.PacketID = 2
	publishOut.Message.Topic = "test"
	publishOut.Message.QOS = 2

	pubrelOut := packet.NewPubrelPacket()
	pubrelOut.PacketID = 2

	publishIn := packet.NewPublishPacket()
	publishIn.PacketID = 1
	publishIn.Message.Topic = "test"
	publishIn.Message.QOS = 2

	pubrecIn := packet.NewPubrecPacket()
	pubrecIn.PacketID = 1

	pubrelIn := packet.NewPubrelPacket()
	pubrelIn.PacketID = 1

	pubcompIn := packet.NewPubcompPacket()
	pubcompIn.PacketID = 1

	disconnect := packet.NewDisconnectPacket()

	port, done := runBroker(t, broker, 2)

	conn1, err := transport.Dial(permittedURL(port))
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	tools.NewFlow().
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
		Test(t, conn1)

	conn2, err := transport.Dial(permittedURL(port))
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(pubrelIn).
		Send(pubcompIn).
		Send(disconnect).
		Close().
		Test(t, conn2)

	<-done
}

func brokerOfflineSubscriptionTest(t *testing.T, broker *Broker, qos uint8) {
	port, done := runBroker(t, broker, 3)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	/* offline subscriber */

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	/* publisher */

	client2 := client.New()

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish("test", []byte("test"), qos, false)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerOfflineSubscriptionRetainedTest(t *testing.T, broker *Broker, qos uint8) {
	port, done := runBroker(t, broker, 3)

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = "test"

	/* offline subscriber */

	client1 := client.New()

	connectFuture1, err := client1.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe("test", qos)
	assert.NoError(t, err)
	assert.NoError(t, subscribeFuture.Wait())
	assert.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	assert.NoError(t, err)

	/* publisher */

	client2 := client.New()

	connectFuture2, err := client2.Connect(permittedURL(port), nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish("test", []byte("test"), qos, true)
	assert.NoError(t, err)
	assert.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	assert.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(qos), msg.QOS)
		assert.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture3.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	assert.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	assert.NoError(t, err)

	<-done
}

func brokerUniqueClientIDTest(t *testing.T, broker *Broker) {
	port, done := runBroker(t, broker, 2)

	options := client.NewOptions()
	options.ClientID = "test"

	wait := make(chan struct{})

	/* first client */

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		assert.Error(t, err)
		close(wait)
	}

	connectFuture1, err := client1.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	/* second client */

	client2 := client.New()

	connectFuture2, err := client2.Connect(permittedURL(port), options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)

	<-done
}
