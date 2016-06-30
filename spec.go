package spec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/require"
)

// A Matrix defines which features should be tested.
type Matrix struct {
	Authentication       bool
	RetainedMessages     bool
	StoredSessions       bool
	StoredSubscriptions  bool
	OfflineSubscriptions bool
	UniqueClientIDs      bool
}

// FullMatrix tests all available features.
var FullMatrix = Matrix{
	Authentication:       true,
	RetainedMessages:     true,
	StoredSessions:       true,
	StoredSubscriptions:  true,
	OfflineSubscriptions: true,
	UniqueClientIDs:      true,
}

// Run will fully test a broker to support all specified features in the matrix.
// The broker being tested should only allow the "allow:allow" login.
func Run(t *testing.T, matrix Matrix, address string) {
	url := fmt.Sprintf("tcp://allow:allow@%s/", address)

	println("Running Broker Publish Subscribe Test (QOS 0)")
	brokerPublishSubscribeTest(t, url, "pubsub/1", "pubsub/1", 0, 0)

	println("Running Broker Publish Subscribe Test (QOS 1)")
	brokerPublishSubscribeTest(t, url, "pubsub/2", "pubsub/2", 1, 1)

	println("Running Broker Publish Subscribe Test (QOS 2)")
	brokerPublishSubscribeTest(t, url, "pubsub/3", "pubsub/3", 2, 2)

	println("Running Broker Publish Subscribe Test (Wildcard One)")
	brokerPublishSubscribeTest(t, url, "pubsub/4/foo", "pubsub/4/+", 0, 0)

	println("Running Broker Publish Subscribe Test (Wildcard Some)")
	brokerPublishSubscribeTest(t, url, "pubsub/5/foo", "pubsub/5/#", 0, 0)

	println("Running Broker Publish Subscribe Test (QOS Downgrade 1->0)")
	brokerPublishSubscribeTest(t, url, "pubsub/6", "pubsub/6", 0, 1)

	println("Running Broker Publish Subscribe Test (QOS Downgrade 2->0)")
	brokerPublishSubscribeTest(t, url, "pubsub/7", "pubsub/7", 0, 2)

	println("Running Broker Publish Subscribe Test (QOS Downgrade 2->1)")
	brokerPublishSubscribeTest(t, url, "pubsub/8", "pubsub/8", 1, 2)

	println("Running Broker Unsubscribe Test (QOS 0)")
	brokerUnsubscribeTest(t, url, "unsub/1", 0)

	println("Running Broker Unsubscribe Test (QOS 1)")
	brokerUnsubscribeTest(t, url, "unsub/2", 1)

	println("Running Broker Unsubscribe Test (QOS 2)")
	brokerUnsubscribeTest(t, url, "unsub/3", 2)

	println("Running Broker Subscription Upgrade Test (QOS 0->1)")
	brokerSubscriptionUpgradeTest(t, url, "subup/1", 0, 1)

	println("Running Broker Subscription Upgrade Test (QOS 1->2)")
	brokerSubscriptionUpgradeTest(t, url, "subup/2", 1, 2)

	println("Running Broker Overlapping Subscriptions Test (Wildcard One)")
	brokerOverlappingSubscriptionsTest(t, url, "ovlsub/foo", "ovlsub/+")

	println("Running Broker Overlapping Subscriptions Test (Wildcard Some)")
	brokerOverlappingSubscriptionsTest(t, url, "ovlsub/foo", "ovlsub/#")

	println("Running Broker Multiple Subscription Test")
	brokerMultipleSubscriptionTest(t, url, "mulsub")

	println("Running Broker Duplicate Subscription Test")
	brokerDuplicateSubscriptionTest(t, url, "dblsub")

	println("Running Broker Will Test (QOS 0)")
	brokerWillTest(t, url, "will/1", 0, 0)

	println("Running Broker Will Test (QOS 1)")
	brokerWillTest(t, url, "will/2", 1, 1)

	println("Running Broker Will Test (QOS 2)")
	brokerWillTest(t, url, "will/3", 2, 2)

	// TODO: Delivers old Wills in case of a crash.

	// TODO: Test Clean Disconnect without forwarding the will.

	if matrix.RetainedMessages {
		println("Running Broker Retained Message Test (QOS 0)")
		brokerRetainedMessageTest(t, url, "retained/1", "retained/1", 0, 0)

		println("Running Broker Retained Message Test (QOS 1)")
		brokerRetainedMessageTest(t, url, "retained/2", "retained/2", 1, 1)

		println("Running Broker Retained Message Test (QOS 2)")
		brokerRetainedMessageTest(t, url, "retained/3", "retained/3", 2, 2)

		println("Running Broker Retained Message Test (Wildcard One)")
		brokerRetainedMessageTest(t, url, "retained/4/foo/bar", "retained/4/foo/+", 0, 0)

		println("Running Broker Retained Message Test (Wildcard Some)")
		brokerRetainedMessageTest(t, url, "retained/5/foo/bar", "retained/5/#", 0, 0)

		println("Running Broker Clear Retained Message Test")
		brokerClearRetainedMessageTest(t, url, "retained/6")

		println("Running Broker Direct Retained Message Test")
		brokerDirectRetainedMessageTest(t, url, "retained/7")

		println("Running Broker Retained Will Test)")
		brokerRetainedWillTest(t, url, "retained/8")
	}

	if matrix.StoredSessions {
		println("Running Broker Publish Resend Test (QOS 1)")
		brokerPublishResendTestQOS1(t, url, "c1", "pubres/1")

		println("Running Broker Publish Resend Test (QOS 2)")
		brokerPublishResendTestQOS2(t, url, "c2", "pubres/2")

		println("Running Broker Pubrel Resend Test (QOS 2)")
		brokerPubrelResendTestQOS2(t, url, "c3", "pubres/3")
	}

	if matrix.StoredSubscriptions {
		println("Running Broker Stored Subscriptions Test (QOS 0)")
		brokerStoredSubscriptionsTest(t, url, "c4", "strdsub/1", 0)

		println("Running Broker Stored Subscriptions Test (QOS 1)")
		brokerStoredSubscriptionsTest(t, url, "c5", "strdsub/2", 1)

		println("Running Broker Stored Subscriptions Test (QOS 2)")
		brokerStoredSubscriptionsTest(t, url, "c6", "strdsub/3", 2)

		println("Running Broker Clean Stored Subscriptions Test")
		brokerCleanStoredSubscriptions(t, url, "c7", "strdsub/4")

		println("Running Broker Remove Stored Subscription Test")
		brokerRemoveStoredSubscription(t, url, "c8", "strdsub/5")
	}

	if matrix.OfflineSubscriptions {
		println("Running Broker Offline Subscription Test (QOS 1)")
		brokerOfflineSubscriptionTest(t, url, "c9", "offsub/1", 1)

		println("Running Broker Offline Subscription Test (QOS 2)")
		brokerOfflineSubscriptionTest(t, url, "c10", "offsub/2", 2)
	}

	if matrix.OfflineSubscriptions && matrix.RetainedMessages {
		println("Running Broker Offline Subscription Test Retained (QOS 1)")
		brokerOfflineSubscriptionRetainedTest(t, url, "c11", "offsubret/1", 1)

		println("Running Broker Offline Subscription Test Retained (QOS 2)")
		brokerOfflineSubscriptionRetainedTest(t, url, "c12", "offsubret/2", 2)
	}

	if matrix.Authentication {
		println("Running Broker Authentication Test")
		brokerAuthenticationTest(t, url)
	}

	if matrix.UniqueClientIDs {
		println("Running Broker Unique Client ID Test")
		brokerUniqueClientIDTest(t, url, "c13")
	}
}

var testPayload = []byte("test")

func brokerPublishSubscribeTest(t *testing.T, url string, out, in string, sub, pub uint8) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, out, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(sub), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(in, sub)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(out, testPayload, pub, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerRetainedMessageTest(t *testing.T, url string, out, in string, sub, pub uint8) {
	require.NoError(t, client.ClearRetainedMessage(url, out))

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	publishFuture, err := client1.Publish(out, testPayload, pub, true)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	err = client1.Disconnect()
	require.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, out, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(sub), msg.QOS)
		require.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe(in, sub)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerClearRetainedMessageTest(t *testing.T, url string, topic string) {
	require.NoError(t, client.ClearRetainedMessage(url, topic))

	// client1 retains message

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	publishFuture1, err := client1.Publish(topic, testPayload, 1, true)
	require.NoError(t, err)
	require.NoError(t, publishFuture1.Wait())

	err = client1.Disconnect()
	require.NoError(t, err)

	// client2 receives retained message and clears it

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(0), msg.QOS)
		require.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	subscribeFuture1, err := client2.Subscribe(topic, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture1.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture1.ReturnCodes)

	<-wait

	publishFuture2, err := client2.Publish(topic, nil, 0, true)
	require.NoError(t, err)
	require.NoError(t, publishFuture2.Wait())

	err = client2.Disconnect()
	require.NoError(t, err)

	// client3 should not receive any message

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		require.Fail(t, "should not be called")
	}

	connectFuture3, err := client3.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture3.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	require.False(t, connectFuture3.SessionPresent)

	subscribeFuture2, err := client3.Subscribe(topic, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture2.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture2.ReturnCodes)

	err = client3.Disconnect()
	require.NoError(t, err)
}

func brokerDirectRetainedMessageTest(t *testing.T, url string, topic string) {
	require.NoError(t, client.ClearRetainedMessage(url, topic))

	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(0), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(topic, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, 0, true)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerWillTest(t *testing.T, url string, topic string, sub, pub uint8) {
	// client1 connects with a will

	client1 := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     pub,
	}

	connectFuture1, err := client1.Connect(url, opts)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	// client2 subscribe to the wills topic

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(sub), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe(topic, sub)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{sub}, subscribeFuture.ReturnCodes)

	// client1 dies

	err = client1.Close()
	require.NoError(t, err)

	// client2 should receive the message

	<-wait

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerRetainedWillTest(t *testing.T, url string, topic string) {
	require.NoError(t, client.ClearRetainedMessage(url, topic))

	// client1 connects with a retained will and dies

	client1 := client.New()

	opts := client.NewOptions()
	opts.Will = &packet.Message{
		Topic:   topic,
		Payload: testPayload,
		QOS:     0,
		Retain:  true,
	}

	connectFuture1, err := client1.Connect(url, opts)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	err = client1.Close()
	require.NoError(t, err)

	// client2 subscribes to the wills topic and receives the retained will

	client2 := client.New()
	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(0), msg.QOS)
		require.True(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	subscribeFuture, err := client2.Subscribe(topic, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	<-wait

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerUnsubscribeTest(t *testing.T, url string, topic string, qos uint8) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic+"/2", msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, qos, msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subscribeFuture, err := client.Subscribe(topic+"/1", qos)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	subscribeFuture, err = client.Subscribe(topic+"/2", qos)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client.Unsubscribe(topic + "/1")
	require.NoError(t, err)
	require.NoError(t, unsubscribeFuture.Wait())

	publishFuture, err := client.Publish(topic+"/1", testPayload, qos, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	publishFuture, err = client.Publish(topic+"/2", testPayload, qos, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerSubscriptionUpgradeTest(t *testing.T, url string, topic string, from, to uint8) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(to), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subscribeFuture1, err := client.Subscribe(topic, from)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture1.Wait())
	require.Equal(t, []uint8{from}, subscribeFuture1.ReturnCodes)

	subscribeFuture2, err := client.Subscribe(topic, to)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture2.Wait())
	require.Equal(t, []uint8{to}, subscribeFuture2.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, to, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerOverlappingSubscriptionsTest(t *testing.T, url string, pub, sub string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, pub, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, byte(0), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subscribeFuture1, err := client.Subscribe(sub, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture1.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture1.ReturnCodes)

	subscribeFuture2, err := client.Subscribe(pub, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture2.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture2.ReturnCodes)

	publishFuture, err := client.Publish(pub, testPayload, 0, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerAuthenticationTest(t *testing.T, url string) {
	deniedURL := strings.Replace(url, "allow:allow", "deny:deny", -1)

	// client1 should be denied

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		require.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture1, err := client1.Connect(deniedURL, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ErrNotAuthorized, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	// client2 should be allowed

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerMultipleSubscriptionTest(t *testing.T, url string, topic string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic+"/3", msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(2), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: topic + "/1", QOS: 0},
		{Topic: topic + "/2", QOS: 1},
		{Topic: topic + "/3", QOS: 2},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{0, 1, 2}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic+"/3", testPayload, 2, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerDuplicateSubscriptionTest(t *testing.T, url string, topic string) {
	client := client.New()
	wait := make(chan struct{})

	client.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(1), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture, err := client.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture.ReturnCode)
	require.False(t, connectFuture.SessionPresent)

	subs := []packet.Subscription{
		{Topic: topic, QOS: 0},
		{Topic: topic, QOS: 1},
	}

	subscribeFuture, err := client.SubscribeMultiple(subs)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{0, 1}, subscribeFuture.ReturnCodes)

	publishFuture, err := client.Publish(topic, testPayload, 1, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client.Disconnect()
	require.NoError(t, err)
}

func brokerStoredSubscriptionsTest(t *testing.T, url string, id, topic string, qos uint8) {
	require.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, qos)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	require.NoError(t, err)

	client2 := client.New()

	wait := make(chan struct{})

	client2.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(qos), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture2, err := client2.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.True(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish(topic, testPayload, qos, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	<-wait

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerCleanStoredSubscriptions(t *testing.T, url string, id, topic string) {
	require.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	require.NoError(t, err)

	client2 := client.New()
	client2.Callback = func(msg *packet.Message, err error) {
		require.Fail(t, "should not be called")
	}

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	publishFuture2, err := client2.Publish(topic, testPayload, 0, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture2.Wait())

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerRemoveStoredSubscription(t *testing.T, url string, id, topic string) {
	require.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, 0)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{0}, subscribeFuture.ReturnCodes)

	unsubscribeFuture, err := client1.Unsubscribe(topic)
	require.NoError(t, err)
	require.NoError(t, unsubscribeFuture.Wait())

	err = client1.Disconnect()
	require.NoError(t, err)

	client2 := client.New()
	client2.Callback = func(msg *packet.Message, err error) {
		require.Fail(t, "should not be called")
	}

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	publishFuture2, err := client2.Publish(topic, testPayload, 0, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture2.Wait())

	err = client2.Disconnect()
	require.NoError(t, err)
}

func brokerPublishResendTestQOS1(t *testing.T, url string, id, topic string) {
	require.NoError(t, client.ClearSession(url, id))

	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = id
	connect.Username = "allow"
	connect.Password = "allow"

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

	conn1, err := transport.Dial(url)
	require.NoError(t, err)
	require.NotNil(t, conn1)

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

	conn2, err := transport.Dial(url)
	require.NoError(t, err)
	require.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(publishIn).
		Send(pubackIn).
		Send(disconnect).
		Close().
		Test(t, conn2)
}

func brokerPublishResendTestQOS2(t *testing.T, url string, id, topic string) {
	require.NoError(t, client.ClearSession(url, id))

	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = id
	connect.Username = "allow"
	connect.Password = "allow"

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

	conn1, err := transport.Dial(url)
	require.NoError(t, err)
	require.NotNil(t, conn1)

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

	conn2, err := transport.Dial(url)
	require.NoError(t, err)
	require.NotNil(t, conn2)

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
}

func brokerPubrelResendTestQOS2(t *testing.T, url string, id, topic string) {
	require.NoError(t, client.ClearSession(url, id))

	connect := packet.NewConnectPacket()
	connect.CleanSession = false
	connect.ClientID = id
	connect.Username = "allow"
	connect.Password = "allow"

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

	conn1, err := transport.Dial(url)
	require.NoError(t, err)
	require.NotNil(t, conn1)

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

	conn2, err := transport.Dial(url)
	require.NoError(t, err)
	require.NotNil(t, conn2)

	publishIn.Dup = true

	tools.NewFlow().
		Send(connect).
		Skip(). // connack
		Receive(pubrelIn).
		Send(pubcompIn).
		Send(disconnect).
		Close().
		Test(t, conn2)
}

func brokerOfflineSubscriptionTest(t *testing.T, url string, id, topic string, qos uint8) {
	require.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	/* offline subscriber */

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, qos)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	require.NoError(t, err)

	/* publisher */

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish(topic, testPayload, qos, false)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	require.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(qos), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture3.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	require.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	require.NoError(t, err)
}

func brokerOfflineSubscriptionRetainedTest(t *testing.T, url string, id, topic string, qos uint8) {
	require.NoError(t, client.ClearSession(url, id))
	require.NoError(t, client.ClearRetainedMessage(url, topic))

	options := client.NewOptions()
	options.CleanSession = false
	options.ClientID = id

	/* offline subscriber */

	client1 := client.New()

	connectFuture1, err := client1.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	subscribeFuture, err := client1.Subscribe(topic, qos)
	require.NoError(t, err)
	require.NoError(t, subscribeFuture.Wait())
	require.Equal(t, []uint8{qos}, subscribeFuture.ReturnCodes)

	err = client1.Disconnect()
	require.NoError(t, err)

	/* publisher */

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, nil)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	publishFuture, err := client2.Publish(topic, testPayload, qos, true)
	require.NoError(t, err)
	require.NoError(t, publishFuture.Wait())

	err = client2.Disconnect()
	require.NoError(t, err)

	/* receiver */

	wait := make(chan struct{})

	client3 := client.New()
	client3.Callback = func(msg *packet.Message, err error) {
		require.NoError(t, err)
		require.Equal(t, topic, msg.Topic)
		require.Equal(t, testPayload, msg.Payload)
		require.Equal(t, uint8(qos), msg.QOS)
		require.False(t, msg.Retain)

		close(wait)
	}

	connectFuture3, err := client3.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture3.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture3.ReturnCode)
	require.True(t, connectFuture3.SessionPresent)

	<-wait

	err = client3.Disconnect()
	require.NoError(t, err)
}

func brokerUniqueClientIDTest(t *testing.T, url string, id string) {
	require.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.ClientID = id

	wait := make(chan struct{})

	/* first client */

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		require.Error(t, err)
		close(wait)
	}

	connectFuture1, err := client1.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture1.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	require.False(t, connectFuture1.SessionPresent)

	/* second client */

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, options)
	require.NoError(t, err)
	require.NoError(t, connectFuture2.Wait())
	require.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	require.False(t, connectFuture2.SessionPresent)

	<-wait

	err = client2.Disconnect()
	require.NoError(t, err)
}
