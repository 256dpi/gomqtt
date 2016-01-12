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
	"fmt"
	"strconv"
	"crypto/tls"
)

var serverTLSConfig *tls.Config
var clientTLSConfig *tls.Config
var testDialer *Dialer

func init() {
	cer, err := tls.LoadX509KeyPair("test/server.pem", "test/server.key")
	if err != nil {
		panic(err)
	}

	serverTLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	clientTLSConfig = &tls.Config{InsecureSkipVerify: true}

	testDialer = NewDialer()
	testDialer.TLSClientConfig = clientTLSConfig
}

// the testPort
type testPort int

// the startPort that gets incremented
var startPort = 55555

// returns a new testPort
func newTestPort() *testPort {
	startPort++
	p := testPort(startPort)
	return &p
}

// generates the listen address for that testPort
func (p *testPort) address() string {
	return fmt.Sprintf("localhost:%d", int(*p))
}

// generates the url for that testPort
func (p *testPort) url(protocol string) string {
	return fmt.Sprintf("%s://localhost:%d/", protocol, int(*p))
}

// return the port as a number
func (p *testPort) port() string {
	return strconv.Itoa(int(*p))
}

// type cast the error to an Error
func toError(err error) Error {
	if terr, ok := err.(Error); ok {
		return terr
	}

	return nil
}
