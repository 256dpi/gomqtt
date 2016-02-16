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

package tools

// Adapted from: https://github.com/phayes/freeport/blob/master/freeport.go.

import (
	"fmt"
	"net"
	"strconv"
)

// A Port is a free TCP port.
type Port int

// NewPort returns a free Port.
func NewPort() *Port {
	// get random free port
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	// use once
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	// defer proper closing
	defer l.Close()

	// return port
	p := Port(l.Addr().(*net.TCPAddr).Port)
	return &p
}

// URL will return the corresponding URL.
func (p *Port) URL(protocol ...string) string {
	proto := "tcp"

	if len(protocol) > 0 {
		proto = protocol[0]
	}

	return fmt.Sprintf("%s://localhost:%d/", proto, int(*p))
}

// Port will return the port number.
func (p *Port) Port() string {
	return strconv.Itoa(int(*p))
}
