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
	//"net"
	"testing"
	//"net/http"
	//"net/url"
	//"time"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/require"
)

func abstractServerTest(t *testing.T, protocol string) {
	tp := newTestPort()

	server, err := testLauncher.Launch(tp.url(protocol))
	require.NoError(t, err)

	go func() {
		conn1, err := server.Accept()
		require.NoError(t, err)

		pkt, err := conn1.Receive()
		require.Equal(t, pkt.Type(), packet.CONNECT)
		require.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		require.NoError(t, err)

		pkt, err = conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	}()

	conn2, err := testDialer.Dial(tp.url(protocol))
	require.NoError(t, err)

	err = conn2.Send(packet.NewConnectPacket())
	require.NoError(t, err)

	pkt, err := conn2.Receive()
	require.Equal(t, pkt.Type(), packet.CONNACK)
	require.NoError(t, err)

	err = conn2.Close()
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)
}

func abstractServerLaunchErrorTest(t *testing.T, protocol string) {
	tp := testPort(1) // <- no permissions

	server, err := testLauncher.Launch(tp.url(protocol))
	require.Equal(t, NetworkError, toError(err).Code())
	require.Nil(t, server)
}

func abstractServerAcceptAfterCloseTest(t *testing.T, protocol string) {
	tp := newTestPort()

	server, err := testLauncher.Launch(tp.url(protocol))
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)

	conn, err := server.Accept()
	require.Nil(t, conn)
	require.Equal(t, NetworkError, toError(err).Code())
}
