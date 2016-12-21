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
	"net"
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

func fakeBroker(t *testing.T, testFlows ...*tools.Flow) (chan struct{}, string) {
	done := make(chan struct{})

	server, err := transport.Launch("tcp://localhost:0")
	assert.NoError(t, err)

	go func() {
		for _, flow := range testFlows {
			conn, err := server.Accept()
			assert.NoError(t, err)

			flow.Test(t, conn)
		}

		err = server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	_, port, _ := net.SplitHostPort(server.Addr().String())

	return done, port
}

func connectPacket() *packet.ConnectPacket {
	pkt := packet.NewConnectPacket()
	pkt.CleanSession = true
	pkt.KeepAlive = 30
	return pkt
}

func connackPacket() *packet.ConnackPacket {
	pkt := packet.NewConnackPacket()
	pkt.ReturnCode = packet.ConnectionAccepted
	pkt.SessionPresent = false
	return pkt
}

func disconnectPacket() *packet.DisconnectPacket {
	return packet.NewDisconnectPacket()
}
