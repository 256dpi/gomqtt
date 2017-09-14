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

package transport

import (
	"crypto/tls"
	"net/url"
)

// The Launcher helps with launching a server and accepting connections.
type Launcher struct {
	TLSConfig *tls.Config
}

// NewLauncher returns a new Launcher.
func NewLauncher() *Launcher {
	return &Launcher{}
}

var sharedLauncher *Launcher

func init() {
	sharedLauncher = NewLauncher()
}

// Launch is a shorthand function.
func Launch(urlString string) (Server, error) {
	return sharedLauncher.Launch(urlString)
}

// Launch will launch a server based on information extracted from an URL.
func (l *Launcher) Launch(urlString string) (Server, error) {
	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, err
	}

	switch urlParts.Scheme {
	case "tcp", "mqtt":
		return NewNetServer(urlParts.Host)
	case "tls", "mqtts":
		return NewSecureNetServer(urlParts.Host, l.TLSConfig)
	case "ws":
		return NewWebSocketServer(urlParts.Host)
	case "wss":
		return NewSecureWebSocketServer(urlParts.Host, l.TLSConfig)
	}

	return nil, ErrUnsupportedProtocol
}
