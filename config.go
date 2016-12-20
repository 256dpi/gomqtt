// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"github.com/gomqtt/packet"
	"github.com/gomqtt/transport"
)

// A Config holds information about establishing a connection to a broker.
type Config struct {
	Dialer       *transport.Dialer
	BrokerURL    string
	ClientID     string
	CleanSession bool
	KeepAlive    string
	WillMessage  *packet.Message
	ValidateSubs bool
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
