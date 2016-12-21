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
	"fmt"
	"net"
)

var serverTLSConfig *tls.Config
var clientTLSConfig *tls.Config

var testDialer *Dialer
var testLauncher *Launcher

func init() {
	cer, err := tls.LoadX509KeyPair("test/server.pem", "test/server.key")
	if err != nil {
		panic(err)
	}

	serverTLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	clientTLSConfig = &tls.Config{InsecureSkipVerify: true}

	testDialer = NewDialer()
	testDialer.TLSConfig = clientTLSConfig

	testLauncher = NewLauncher()
	testLauncher.TLSConfig = serverTLSConfig
}

// returns a client-ish and server-ish pair of connections
func connectionPair(protocol string, handler func(Conn)) (Conn, chan struct{}) {
	done := make(chan struct{})

	server, err := testLauncher.Launch(protocol + "://localhost:0")
	if err != nil {
		panic(err)
	}

	go func() {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}

		handler(conn)

		server.Close()
		close(done)
	}()

	conn, err := testDialer.Dial(getURL(server, protocol))
	if err != nil {
		panic(err)
	}

	return conn, done
}

func getPort(s Server) string {
	_, port, _ := net.SplitHostPort(s.Addr().String())
	return port
}

func getURL(s Server, protocol string) string {
	return fmt.Sprintf("%s://%s", protocol, s.Addr().String())
}
