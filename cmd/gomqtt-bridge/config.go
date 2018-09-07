package main

import (
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

// Subscription is a single subscription in a Subscribe packet.
type subscription struct {
	Topic string `json:"topic" validate:"nonzero"`
	QOS   uint8  `json:"qos" default:"0"`
}

// bridge configuration
type bridge struct {
	Local  config `json:"local" validate:"nonzero"`
	Remote config `json:"remote" validate:"nonzero"`
}

// mqtt client configuration
type config struct {
	Address       string         `json:"address" validate:"nonzero"`
	ClientID      string         `json:"clientid" validate:"nonzero"`
	KeepAlive     time.Duration  `json:"keepalive" default:"30s"`
	Subscriptions []subscription `json:"subscriptions"`
}

// convert to client config
func (c *config) toClientConfig() *client.Config {
	cc := client.NewConfig(c.Address)
	cc.ClientID = c.ClientID
	cc.CleanSession = false
	if c.KeepAlive != 0 {
		cc.KeepAlive = c.KeepAlive.String()
	}
	return cc
}

func toSubscriptions(subs []subscription) []packet.Subscription {
	output := make([]packet.Subscription, len(subs))
	for i, s := range subs {
		output[i] = packet.Subscription{Topic: s.Topic, QOS: s.QOS}
	}
	return output
}
