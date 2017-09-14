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
	"io"
	"testing"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func abstractServerTest(t *testing.T, protocol string) {
	server, err := testLauncher.Launch(protocol + "://localhost:0")
	require.NoError(t, err)

	go func() {
		conn1, err := server.Accept()
		require.NoError(t, err)

		pkt, err := conn1.Receive()
		assert.Equal(t, pkt.Type(), packet.CONNECT)
		assert.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		assert.NoError(t, err)

		pkt, err = conn1.Receive()
		assert.Nil(t, pkt)
		assert.Equal(t, io.EOF, err)
	}()

	conn2, err := testDialer.Dial(getURL(server, protocol))
	require.NoError(t, err)

	err = conn2.Send(packet.NewConnectPacket())
	assert.NoError(t, err)

	pkt, err := conn2.Receive()
	assert.Equal(t, pkt.Type(), packet.CONNACK)
	assert.NoError(t, err)

	err = conn2.Close()
	assert.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func abstractServerLaunchErrorTest(t *testing.T, protocol string) {
	server, err := testLauncher.Launch(protocol + "://localhost:1")
	assert.Error(t, err)
	assert.Nil(t, server)
}

func abstractServerAcceptAfterCloseTest(t *testing.T, protocol string) {
	server, err := testLauncher.Launch(protocol + "://localhost:0")
	require.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)

	conn, err := server.Accept()
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func abstractServerCloseAfterCloseTest(t *testing.T, protocol string) {
	server, err := testLauncher.Launch(protocol + "://localhost:0")
	require.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)

	err = server.Close()
	assert.Error(t, err)
}

func abstractServerAddrTest(t *testing.T, protocol string) {
	server, err := testLauncher.Launch(protocol + "://localhost:0")
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("127.0.0.1:%s", getPort(server)), server.Addr().String())

	err = server.Close()
	assert.NoError(t, err)
}
