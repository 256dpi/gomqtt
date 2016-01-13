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
	"net"
	"net/http"
	"net/url"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

var ErrUnsupportedProtocol = errors.New("dialer: unsupported protocol")

type Dialer struct {
	TLSConfig *tls.Config
	RequestHeader http.Header

	DefaultTCPPort string
	DefaultTLSPort string
	DefaultWSPort  string
	DefaultWSSPort string

	webSocketDialer *websocket.Dialer
}

func NewDialer() *Dialer {
	return &Dialer{
		DefaultTCPPort: "1883",
		DefaultTLSPort: "8883",
		DefaultWSPort:  "80",
		DefaultWSSPort: "443",
		webSocketDialer: &websocket.Dialer{
			Proxy: http.ProxyFromEnvironment,
			Subprotocols: []string{ "mqtt" },
		},
	}
}

var sharedDialer *Dialer

func init() {
	sharedDialer = NewDialer()
}

func Dial(urlString string) (Conn, error) {
	return sharedDialer.Dial(urlString)
}

func (d *Dialer) Dial(urlString string) (Conn, error) {
	urlParts, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, newTransportError(DialError, err)
	}

	host, port, err := net.SplitHostPort(urlParts.Host)
	if err != nil {
		host = urlParts.Host
		port = ""
	}

	switch urlParts.Scheme {
	case "mqtt", "tcp":
		if port == "" {
			port = d.DefaultTCPPort
		}

		conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return nil, newTransportError(DialError, err)
		}

		return NewNetConn(conn), nil
	case "mqtts", "tcps", "ssl", "tls":
		if port == "" {
			port = d.DefaultTLSPort
		}

		conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), d.TLSConfig)
		if err != nil {
			return nil, newTransportError(DialError, err)
		}

		return NewNetConn(conn), nil
	case "ws":
		if port == "" {
			port = d.DefaultWSPort
		}

		url := fmt.Sprintf("ws://%s:%s%s", host, port, urlParts.Path)

		conn, _, err := d.webSocketDialer.Dial(url, d.RequestHeader)
		if err != nil {
			return nil, newTransportError(DialError, err)
		}

		return NewWebSocketConn(conn), nil
	case "wss":
		if port == "" {
			port = d.DefaultWSSPort
		}

		url := fmt.Sprintf("wss://%s:%s%s", host, port, urlParts.Path)

		d.webSocketDialer.TLSClientConfig = d.TLSConfig
		conn, _, err := d.webSocketDialer.Dial(url, d.RequestHeader)
		if err != nil {
			return nil, newTransportError(DialError, err)
		}

		return NewWebSocketConn(conn), nil
	}

	return nil, newTransportError(DialError, ErrUnsupportedProtocol)
}
