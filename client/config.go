package client

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
)

// A Config holds information about establishing a connection to a broker.
type Config struct {
	Dialer       *transport.Dialer
	BrokerURL    string
	ClientID     string
	Username     string
	Password     string
	CleanSession bool
	KeepAlive    string
	WillMessage  *packet.Message
	ValidateSubs bool

	// Custom methods
	// Unless you understand these behaviors, do not change them.
	ProcessPublish func(*packet.Publish) error
	ProcessPuback  func(*packet.Puback) error
	ProcessPubcomp func(*packet.Pubcomp) error
	ProcessPubrec  func(*packet.Pubrec) error
	ProcessPubrel  func(*packet.Pubrel) error
}

// NewConfig creates a new Config using the specified URL.
func NewConfig(url string) *Config {
	return &Config{
		BrokerURL:    url,
		CleanSession: true,
		KeepAlive:    "30s",
		ValidateSubs: true,
	}
}

// NewConfigWithClientID creates a new Config using the specified URL and client ID.
func NewConfigWithClientID(url, id string) *Config {
	config := NewConfig(url)
	config.ClientID = id
	return config
}
