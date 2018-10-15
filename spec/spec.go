// Package spec implements a reusable specification test for MQTT brokers.
package spec

import (
	"fmt"
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

	counter int
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
	uri, err := url.Parse(c.URL)
	if err != nil {
		panic(err)
	}

	if uri.User == nil {
		return "", ""
	}

	pw, _ := uri.User.Password()

	return uri.User.Username(), pw
}

func (c *Config) clientID() string {
	c.counter++
	return fmt.Sprintf("c%d", c.counter)
}

// Run will fully test a to support all specified features in the matrix.
func Run(t *testing.T, config *Config) {
	t.Run("PublishSubscribeQOS00", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/1", "pubsub/1", 0, 0, 0)
	})

	t.Run("PublishSubscribeQOS11", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/2", "pubsub/2", 1, 1, 1)
	})

	t.Run("PublishSubscribeQOS22", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/3", "pubsub/3", 2, 2, 2)
	})

	t.Run("PublishSubscribeWildcardOne", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/4/foo", "pubsub/4/+", 0, 0, 0)
	})

	t.Run("PublishSubscribeWildcardSome", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/5/foo", "pubsub/5/#", 0, 0, 0)
	})

	t.Run("PublishSubscribeQOSDowngrade1To0", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/6", "pubsub/6", 0, 1, 0)
	})

	t.Run("PublishSubscribeQOSDowngrade2To0", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/7", "pubsub/7", 0, 2, 0)
	})

	t.Run("PublishSubscribeQOSDowngrade2To1", func(t *testing.T) {
		PublishSubscribeTest(t, config, "pubsub/8", "pubsub/8", 1, 2, 1)
	})

	t.Run("UnsubscribeQOS0", func(t *testing.T) {
		UnsubscribeTest(t, config, "unsub/1", 0)
	})

	t.Run("UnsubscribeQOS1", func(t *testing.T) {
		UnsubscribeTest(t, config, "unsub/2", 1)
	})

	t.Run("UnsubscribeQOS2", func(t *testing.T) {
		UnsubscribeTest(t, config, "unsub/3", 2)
	})

	t.Run("UnsubscribeNotExistingSubscription", func(t *testing.T) {
		UnsubscribeNotExistingSubscriptionTest(t, config, "unsub/4")
	})

	t.Run("UnsubscribeOverlappingSubscription", func(t *testing.T) {
		UnsubscribeOverlappingSubscriptions(t, config, "unsub/5")
	})

	t.Run("SubscriptionUpgradeQOS0To1", func(t *testing.T) {
		SubscriptionUpgradeTest(t, config, "subup/1", 0, 1)
	})

	t.Run("SubscriptionUpgradeQOS1To2", func(t *testing.T) {
		SubscriptionUpgradeTest(t, config, "subup/2", 1, 2)
	})

	t.Run("OverlappingSubscriptionsWildcardOne", func(t *testing.T) {
		OverlappingSubscriptionsTest(t, config, "ovlsub/1/foo", "ovlsub/1/+")
	})

	t.Run("OverlappingSubscriptionsWildcardSome", func(t *testing.T) {
		OverlappingSubscriptionsTest(t, config, "ovlsub/2/foo", "ovlsub/2/#")
	})

	t.Run("MultipleSubscription", func(t *testing.T) {
		MultipleSubscriptionTest(t, config, "mulsub")
	})

	t.Run("DuplicateSubscription", func(t *testing.T) {
		DuplicateSubscriptionTest(t, config, "dblsub")
	})

	t.Run("IsolatedSubscription", func(t *testing.T) {
		IsolatedSubscriptionTest(t, config, "islsub")
	})

	t.Run("WillQOS0", func(t *testing.T) {
		WillTest(t, config, "will/1", 0, 0)
	})

	t.Run("WillQOS1", func(t *testing.T) {
		WillTest(t, config, "will/2", 1, 1)
	})

	t.Run("WillQOS2", func(t *testing.T) {
		WillTest(t, config, "will/3", 2, 2)
	})

	t.Run("CleanWill", func(t *testing.T) {
		CleanWillTest(t, config, "will/4")
	})

	t.Run("KeepAlive", func(t *testing.T) {
		KeepAliveTest(t, config)
	})

	t.Run("KeepAliveTimeout", func(t *testing.T) {
		KeepAliveTimeoutTest(t, config)
	})

	if config.RetainedMessages {
		t.Run("RetainedMessageQOS0", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/1", "retained/1", 0, 0)
		})

		t.Run("RetainedMessageQOS1", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/2", "retained/2", 1, 1)
		})

		t.Run("RetainedMessageQOS2", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/3", "retained/3", 2, 2)
		})

		t.Run("RetainedMessageDowngrade1To0", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/4", "retained/4", 0, 1)
		})

		t.Run("RetainedMessageDowngrade2To0", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/5", "retained/5", 0, 2)
		})

		t.Run("RetainedMessageDowngrade2To1", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/6", "retained/6", 1, 2)
		})

		t.Run("RetainedMessageWildcardOne", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/7/foo/bar", "retained/7/foo/+", 0, 0)
		})

		t.Run("RetainedMessageWildcardSome", func(t *testing.T) {
			RetainedMessageTest(t, config, "retained/8/foo/bar", "retained/8/#", 0, 0)
		})

		t.Run("RetainedMessageReplace", func(t *testing.T) {
			RetainedMessageReplaceTest(t, config, "retained/9")
		})

		t.Run("ClearRetainedMessage", func(t *testing.T) {
			ClearRetainedMessageTest(t, config, "retained/10")
		})

		t.Run("DirectRetainedMessage", func(t *testing.T) {
			DirectRetainedMessageTest(t, config, "retained/11")
		})

		t.Run("DirectClearRetainedMessage", func(t *testing.T) {
			DirectClearRetainedMessageTest(t, config, "retained/12")
		})

		t.Run("RetainedWill", func(t *testing.T) {
			RetainedWillTest(t, config, "retained/13")
		})

		t.Run("RetainedMessageResubscription", func(t *testing.T) {
			RetainedMessageResubscriptionTest(t, config, "retained/14")
		})
	}

	if config.StoredPackets {
		t.Run("PublishResendQOS1", func(t *testing.T) {
			PublishResendQOS1Test(t, config, "pubres/1")
		})

		t.Run("PublishResendQOS2", func(t *testing.T) {
			PublishResendQOS2Test(t, config, "pubres/2")
		})

		t.Run("PubrelResendQOS2", func(t *testing.T) {
			PubrelResendQOS2Test(t, config, "pubres/3")
		})
	}

	if config.StoredSubscriptions {
		t.Run("StoredSubscriptionsQOS0", func(t *testing.T) {
			StoredSubscriptionsTest(t, config, "strdsub/1", 0)
		})

		t.Run("StoredSubscriptionsQOS1", func(t *testing.T) {
			StoredSubscriptionsTest(t, config, "strdsub/2", 1)
		})

		t.Run("StoredSubscriptionsQOS2", func(t *testing.T) {
			StoredSubscriptionsTest(t, config, "strdsub/3", 2)
		})

		t.Run("CleanStoredSubscriptions", func(t *testing.T) {
			CleanStoredSubscriptionsTest(t, config, "strdsub/4")
		})

		t.Run("RemoveStoredSubscription", func(t *testing.T) {
			RemoveStoredSubscriptionTest(t, config, "strdsub/5")
		})
	}

	if config.OfflineSubscriptions {
		t.Run("OfflineSubscriptionQOS00", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/1", 0, 0, false)
		})

		t.Run("OfflineSubscriptionQOS01", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/2", 0, 1, false)
		})

		t.Run("OfflineSubscriptionQOS10", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/3", 1, 0, false)
		})

		t.Run("OfflineSubscriptionQOS11", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/4", 1, 1, true)
		})

		t.Run("OfflineSubscriptionQOS12", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/5", 1, 2, true)
		})

		t.Run("OfflineSubscriptionQOS21", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/6", 2, 1, true)
		})

		t.Run("OfflineSubscriptionQOS22", func(t *testing.T) {
			OfflineSubscriptionTest(t, config, "offsub/7", 2, 2, true)
		})
	}

	if config.OfflineSubscriptions && config.RetainedMessages {
		t.Run("OfflineSubscriptionRetainedQOS0", func(t *testing.T) {
			OfflineSubscriptionRetainedTest(t, config, "offsubret/1", 0, 0, false)
		})

		t.Run("OfflineSubscriptionRetainedQOS1", func(t *testing.T) {
			OfflineSubscriptionRetainedTest(t, config, "offsubret/2", 1, 1, true)
		})

		t.Run("OfflineSubscriptionRetainedQOS2", func(t *testing.T) {
			OfflineSubscriptionRetainedTest(t, config, "offsubret/3", 2, 2, true)
		})
	}

	if config.Authentication {
		t.Run("Authentication", func(t *testing.T) {
			AuthenticationTest(t, config)
		})
	}

	if config.UniqueClientIDs {
		t.Run("UniqueClientIDUnclean", func(t *testing.T) {
			UniqueClientIDUncleanTest(t, config)
		})

		t.Run("UniqueClientIDClean", func(t *testing.T) {
			UniqueClientIDCleanTest(t, config)
		})
	}

	if config.RootSlashDistinction {
		t.Run("RootSlashDistinction", func(t *testing.T) {
			RootSlashDistinctionTest(t, config, "rootslash")
		})
	}
}
