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

import "github.com/gomqtt/packet"

// Options are passed to a Client on Connect.
type Options struct {
	BrokerURL    string
	ClientID     string
	CleanSession bool
	KeepAlive    string
	WillMessage  *packet.Message
	ValidateSubs bool
}

// NewOptions will initialize and return a new Options struct.
func NewOptions(url string) *Options {
	return &Options{
		BrokerURL:    url,
		CleanSession: true,
		KeepAlive:    "30s",
		ValidateSubs: true,
	}
}

// NewOptionsWithClientID will initialize and return a new Options struct.
func NewOptionsWithClientID(url string, id string) *Options {
	opts := NewOptions(url)
	opts.ClientID = id
	return opts
}
