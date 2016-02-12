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

package broker

import (
	"net"
	"fmt"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
"testing"
)

// the testPort
type testPort int

// returns a new testPort
func newTestPort() *testPort {
	// taken from: https://github.com/phayes/freeport/blob/master/freeport.go

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	p := testPort(l.Addr().(*net.TCPAddr).Port)
	return &p
}

// generates the url for that testPort
func (p *testPort) url() string {
	return fmt.Sprintf("tcp://localhost:%d/", int(*p))
}

// generates a protected url for that testPort
func (p *testPort) protectedURL(user, password string) string {
	return fmt.Sprintf("tcp://%s:%s@localhost:%d/", user, password, int(*p))
}

func startBroker(t *testing.T, broker *Broker, num int) (*testPort, chan struct{}) {
	tp := newTestPort()

	server, err := transport.Launch(tp.url())
	assert.NoError(t, err)

	done := make(chan struct{})

	go func(){
		for i:=0; i<num; i++ {
			conn, err := server.Accept()
			assert.NoError(t, err)

			broker.Handle(conn)
		}

		err := server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	return tp, done
}
