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
	"crypto/tls"
	"fmt"
	"github.com/gomqtt/stream"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
)

var sharedWebSocketDialer = &websocket.Dialer{
	Proxy: http.ProxyFromEnvironment,
	Subprotocols: []string { "mqtt", "mqttv3.1" },
}

func dial(opts *Options) (stream.Stream, error) {
	host, port, err := net.SplitHostPort(opts.URL.Host)
	if err != nil {
		return nil, err
	}

	switch opts.URL.Scheme {
	case "mqtt", "tcp":
		if port == "" {
			port = "1883"
		}

		conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, err
		}

		return stream.NewNetStream(conn), nil
	case "mqtts", "tcps", "ssl", "tls":
		if port == "" {
			port = "8883"
		}

		// TODO: SSL Config
		conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), nil)
		if err != nil {
			return nil, err
		}

		return stream.NewNetStream(conn), nil
	case "ws":
		if port == "" {
			port = "80"
		}

		url := fmt.Sprintf("ws://%s:%s/", host, port)

		//TODO: WebSocket Config
		conn, _, err := sharedWebSocketDialer.Dial(url, nil)
		if err != nil {
			return nil, err
		}

		return stream.NewWebSocketStream(conn), nil
	case "wss":
		if port == "" {
			port = "443"
		}

		url := fmt.Sprintf("ws://%s:%s/", host, port)

		//TODO: Secure WebSocket Config
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			return nil, err
		}

		return stream.NewWebSocketStream(conn), nil
	}

	return nil, fmt.Errorf("Undetectable Protocol")
}
