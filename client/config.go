package client

import (
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

// Dialer defines the dialer used by a client.
type Dialer interface {
	Dial(urlString string) (transport.Conn, error)
}

// A Config holds information about establishing a connection to a broker.
type Config struct {
	// Dialer can be set to use a custom dialer.
	Dialer Dialer

	// BrokerURL is the url that is used to infer options to open the connection.
	BrokerURL string

	// ClientID can be set to the client's id.
	ClientID string

	// CleanSession can be set to request a clean session.
	CleanSession bool

	// KeepAlive should be time a duration string e.g. "30s".
	KeepAlive string

	// Will message is registered on the broker upon connect if set.
	WillMessage *packet.Message

	// ValidateSubs will cause the client to fail if subscriptions failed.
	ValidateSubs bool

	// ReadLimit defines the maximum size of a packet that can be received.
	ReadLimit int64

	// MaxWriteDelay defines the maximum allowed delay when flushing the
	// underlying buffered writer.
	MaxWriteDelay time.Duration

	// AlwaysAnnounceOnPublish defines when the message callback handler is called.
	// QOS 0,1: callback always occurs on reception of Publish
	// QOS 2 and AlwaysAnnounceOnPublish == false: callback occurs on reception of PubRel and returning and error in
	//   the callback will close the connection before PubComp is sent.
	// QOS 2 and AlwaysAnnounceOnPublish == true: callback occurs on reception of Publish and returning an error in
	//   the callback will close the connection, preventing PubRec being sent, thus ensuring redelivery of publish.
	AlwaysAnnounceOnPublish bool
}

// NewConfig creates a new Config using the specified URL.
func NewConfig(url string) *Config {
	return &Config{
		BrokerURL:     url,
		CleanSession:  true,
		KeepAlive:     "30s",
		ValidateSubs:  true,
		ReadLimit:     8 * 1024 * 1024, // 8MB
		MaxWriteDelay: 10 * time.Millisecond,
	}
}

// NewConfigWithClientID creates a new Config using the specified URL and client ID.
func NewConfigWithClientID(url, id string) *Config {
	config := NewConfig(url)
	config.ClientID = id
	return config
}
