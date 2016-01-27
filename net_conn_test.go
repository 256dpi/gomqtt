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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gomqtt/packet"
)

func TestNetConnConnection(t *testing.T) {
	abstractConnConnectTest(t, "tcp")
}

func TestNetConnClose(t *testing.T) {
	abstractConnCloseTest(t, "tcp")
}

func TestNetConnEncodeError(t *testing.T) {
	abstractConnEncodeErrorTest(t, "tcp")
}

func TestNetConnDecode1Error(t *testing.T) {
	abstractConnDecodeError1Test(t, "tcp")
}

func TestNetConnDecode2Error(t *testing.T) {
	abstractConnDecodeError2Test(t, "tcp")
}

func TestNetConnDecode3Error(t *testing.T) {
	abstractConnDecodeError3Test(t, "tcp")
}

func TestNetConnSendAfterClose(t *testing.T) {
	abstractConnSendAfterCloseTest(t, "tcp")
}

func TestNetConnCounters(t *testing.T) {
	abstractConnCountersTest(t, "tcp")
}

func TestNetConnReadLimit(t *testing.T) {
	abstractConnReadLimitTest(t, "tcp")
}

func TestNetConnCloseAfterClose(t *testing.T) {
	abstractConnCloseAfterCloseTest(t, "tcp")
}

func TestNetConnCloseWhileReadError(t *testing.T) {
	conn2, done := abstractConnTestPreparer("tcp", func(conn1 Conn) {
		pkt := packet.NewPublishPacket()
		pkt.Topic = []byte("foo/bar/baz")
		buf := make([]byte, pkt.Len())
		pkt.Encode(buf)

		netConn := conn1.(*NetConn)
		_, err := netConn.conn.Write(buf[0:7]) // <- incomplete packet
		assert.NoError(t, err)

		err = netConn.conn.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, NetworkError, toError(err).Code())

	<-done
}

func TestNetConnCloseWhileDetectError(t *testing.T) {
	conn2, done := abstractConnTestPreparer("tcp", func(conn1 Conn) {
		pkt := packet.NewPublishPacket()
		pkt.Topic = []byte("foo/bar/baz")
		buf := make([]byte, pkt.Len())
		pkt.Encode(buf)

		netConn := conn1.(*NetConn)
		_, err := netConn.conn.Write(buf[0:1]) // <- too less for a detection
		assert.NoError(t, err)

		err = netConn.conn.Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Equal(t, NetworkError, toError(err).Code())

	<-done
}
