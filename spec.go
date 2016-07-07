package spec

import (
	"net/url"
	"testing"
	"time"
)

var testPayload = []byte("test")
var testPayload2 = []byte("test2")

// A Config defines which features should be tested.
type Config struct {
	URL     string
	DenyURL string

	RetainedMessages     bool
	StoredPackets        bool
	StoredSubscriptions  bool
	OfflineSubscriptions bool
	Authentication       bool
	UniqueClientIDs      bool
	RootSlashDistinction bool

	// ProcessWait defines the time some tests should wait and let the broker
	// finish processing (e.g. properly terminating a connection)
	ProcessWait time.Duration

	// MessageRetainWait defines the time retain test should wait to be sure
	// the published messages has been retained.
	MessageRetainWait time.Duration

	// NoMessageWait defines the time some tests should wait for eventually
	// receiving a wrongly sent message or an error.
	NoMessageWait time.Duration
}

// AllFeatures returns a config that enables all features.
func AllFeatures() *Config {
	return &Config{
		RetainedMessages:     true,
		StoredPackets:        true,
		StoredSubscriptions:  true,
		OfflineSubscriptions: true,
		Authentication:       true,
		UniqueClientIDs:      true,
		RootSlashDistinction: true,
	}
}

func (c *Config) usernamePassword() (string, string) {
	url, err := url.Parse(c.URL)
	if err != nil {
		panic(err)
	}

	if url.User == nil {
		return "", ""
	}

	pw, _ := url.User.Password()

	return url.User.Username(), pw
}

// Run will fully test a to support all specified features in the matrix.
func Run(t *testing.T, config *Config) {
	println("Running Publish Subscribe Test (QOS 0)")
	PublishSubscribeTest(t, config, "pubsub/1", "pubsub/1", 0, 0)

	println("Running Publish Subscribe Test (QOS 1)")
	PublishSubscribeTest(t, config, "pubsub/2", "pubsub/2", 1, 1)

	println("Running Publish Subscribe Test (QOS 2)")
	PublishSubscribeTest(t, config, "pubsub/3", "pubsub/3", 2, 2)

	println("Running Publish Subscribe Test (Wildcard One)")
	PublishSubscribeTest(t, config, "pubsub/4/foo", "pubsub/4/+", 0, 0)

	println("Running Publish Subscribe Test (Wildcard Some)")
	PublishSubscribeTest(t, config, "pubsub/5/foo", "pubsub/5/#", 0, 0)

	println("Running Publish Subscribe Test (QOS Downgrade 1->0)")
	PublishSubscribeTest(t, config, "pubsub/6", "pubsub/6", 0, 1)

	println("Running Publish Subscribe Test (QOS Downgrade 2->0)")
	PublishSubscribeTest(t, config, "pubsub/7", "pubsub/7", 0, 2)

	println("Running Publish Subscribe Test (QOS Downgrade 2->1)")
	PublishSubscribeTest(t, config, "pubsub/8", "pubsub/8", 1, 2)

	println("Running Unsubscribe Test (QOS 0)")
	UnsubscribeTest(t, config, "unsub/1", 0)

	println("Running Unsubscribe Test (QOS 1)")
	UnsubscribeTest(t, config, "unsub/2", 1)

	println("Running Unsubscribe Test (QOS 2)")
	UnsubscribeTest(t, config, "unsub/3", 2)

	println("Running Unsubscribe Not Existing Subscription Test")
	UnsubscribeNotExistingSubscriptionTest(t, config, "unsub/4")

	println("Running Unsubscribe Overlapping Subscription Test")
	UnsubscribeOverlappingSubscriptions(t, config, "unsub/5")

	println("Running Subscription Upgrade Test (QOS 0->1)")
	SubscriptionUpgradeTest(t, config, "subup/1", 0, 1)

	println("Running Subscription Upgrade Test (QOS 1->2)")
	SubscriptionUpgradeTest(t, config, "subup/2", 1, 2)

	println("Running Overlapping Subscriptions Test (Wildcard One)")
	OverlappingSubscriptionsTest(t, config, "ovlsub/1/foo", "ovlsub/1/+")

	println("Running Overlapping Subscriptions Test (Wildcard Some)")
	OverlappingSubscriptionsTest(t, config, "ovlsub/2/foo", "ovlsub/2/#")

	println("Running Multiple Subscription Test")
	MultipleSubscriptionTest(t, config, "mulsub")

	println("Running Duplicate Subscription Test")
	DuplicateSubscriptionTest(t, config, "dblsub")

	println("Running Isolated Subscription Test")
	IsolatedSubscriptionTest(t, config, "islsub")

	println("Running Will Test (QOS 0)")
	WillTest(t, config, "will/1", 0, 0)

	println("Running Will Test (QOS 1)")
	WillTest(t, config, "will/2", 1, 1)

	println("Running Will Test (QOS 2)")
	WillTest(t, config, "will/3", 2, 2)

	println("Running Clean Will Test")
	CleanWillTest(t, config, "will/4")

	println("Running Keep Alive Test")
	KeepAliveTest(t, config)

	println("Running Keep Alive Timeout Test")
	KeepAliveTimeoutTest(t, config)

	if config.RetainedMessages {
		println("Running Retained Message Test (QOS 0)")
		RetainedMessageTest(t, config, "retained/1", "retained/1", 0, 0)

		println("Running Retained Message Test (QOS 1)")
		RetainedMessageTest(t, config, "retained/2", "retained/2", 1, 1)

		println("Running Retained Message Test (QOS 2)")
		RetainedMessageTest(t, config, "retained/3", "retained/3", 2, 2)

		println("Running Retained Message Test (Wildcard One)")
		RetainedMessageTest(t, config, "retained/4/foo/bar", "retained/4/foo/+", 0, 0)

		println("Running Retained Message Test (Wildcard Some)")
		RetainedMessageTest(t, config, "retained/5/foo/bar", "retained/5/#", 0, 0)

		println("Running Retained Message Replace Test")
		RetainedMessageReplaceTest(t, config, "retained/6")

		println("Running Clear Retained Message Test")
		ClearRetainedMessageTest(t, config, "retained/7")

		println("Running Direct Retained Message Test")
		DirectRetainedMessageTest(t, config, "retained/8")

		println("Running Retained Will Test")
		RetainedWillTest(t, config, "retained/9")
	}

	if config.StoredPackets {
		println("Running Publish Resend Test (QOS 1)")
		PublishResendQOS1Test(t, config, "c1", "pubres/1")

		println("Running Publish Resend Test (QOS 2)")
		PublishResendQOS2Test(t, config, "c2", "pubres/2")

		println("Running Pubrel Resend Test (QOS 2)")
		PubrelResendQOS2Test(t, config, "c3", "pubres/3")
	}

	if config.StoredSubscriptions {
		println("Running Stored Subscriptions Test (QOS 0)")
		StoredSubscriptionsTest(t, config, "c4", "strdsub/1", 0)

		println("Running Stored Subscriptions Test (QOS 1)")
		StoredSubscriptionsTest(t, config, "c5", "strdsub/2", 1)

		println("Running Stored Subscriptions Test (QOS 2)")
		StoredSubscriptionsTest(t, config, "c6", "strdsub/3", 2)

		println("Running Clean Stored Subscriptions Test")
		CleanStoredSubscriptionsTest(t, config, "c7", "strdsub/4")

		println("Running Remove Stored Subscription Test")
		RemoveStoredSubscriptionTest(t, config, "c8", "strdsub/5")
	}

	if config.OfflineSubscriptions {
		println("Running Offline Subscription Test (QOS 1)")
		OfflineSubscriptionTest(t, config, "c9", "offsub/1", 1)

		println("Running Offline Subscription Test (QOS 2)")
		OfflineSubscriptionTest(t, config, "c10", "offsub/2", 2)
	}

	if config.OfflineSubscriptions && config.RetainedMessages {
		println("Running Offline Subscription Retained Test (QOS 1)")
		OfflineSubscriptionRetainedTest(t, config, "c11", "offsubret/1", 1)

		println("Running Offline Subscription Retained Test (QOS 2)")
		OfflineSubscriptionRetainedTest(t, config, "c12", "offsubret/2", 2)
	}

	if config.Authentication {
		println("Running Authentication Test")
		AuthenticationTest(t, config)
	}

	if config.UniqueClientIDs {
		println("Running Unique Client ID Test")
		UniqueClientIDTest(t, config, "c13")
	}

	if config.RootSlashDistinction {
		println("Running Root Slash Distinction Test")
		RootSlashDistinctionTest(t, config, "rootslash")
	}
}
