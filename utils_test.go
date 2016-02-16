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
	"testing"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func errorCallback(t *testing.T) func(*packet.Message, error) {
	return func(msg *packet.Message, err error) {
		if err != nil {
			println(err.Error())
		}

		assert.Fail(t, "callback should not have been called")
	}
}

func startBroker(t *testing.T, broker *Broker, num int) (*tools.Port, chan struct{}) {
	port := tools.NewPort()

	server, err := transport.Launch(port.URL())
	assert.NoError(t, err)

	done := make(chan struct{})

	go func() {
		for i := 0; i < num; i++ {
			conn, err := server.Accept()
			assert.NoError(t, err)

			broker.Handle(conn)
		}

		err := server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	return port, done
}
