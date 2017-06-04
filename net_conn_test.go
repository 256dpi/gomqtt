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
	"time"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
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

func TestNetConnDecodeError(t *testing.T) {
	abstractConnDecodeErrorTest(t, "tcp")
}

func TestNetConnSendAfterClose(t *testing.T) {
	abstractConnSendAfterCloseTest(t, "tcp")
}

func TestNetConnCloseWhileSend(t *testing.T) {
	abstractConnCloseWhileSendTest(t, "tcp")
}

func TestNetConnSendAndCloseTest(t *testing.T) {
	abstractConnSendAndCloseTest(t, "tcp")
}

func TestNetConnReadLimit(t *testing.T) {
	abstractConnReadLimitTest(t, "tcp")
}

func TestNetConnReadTimeout(t *testing.T) {
	abstractConnReadTimeoutTest(t, "tcp")
}

func TestNetConnCloseAfterClose(t *testing.T) {
	abstractConnCloseAfterCloseTest(t, "tcp")
}

func TestNetConnAddr(t *testing.T) {
	abstractConnAddrTest(t, "tcp")
}

func TestNetConnBufferedSend(t *testing.T) {
	abstractConnBufferedSendTest(t, "tcp")
}

func TestNetConnSendAfterBufferedSend(t *testing.T) {
	abstractConnSendAfterBufferedSendTest(t, "tcp")
}

func TestNetConnBufferedSendAfterClose(t *testing.T) {
	abstractConnBufferedSendAfterCloseTest(t, "tcp")
}

func TestNetConnCloseAfterBufferedSend(t *testing.T) {
	abstractConnCloseAfterBufferedSendTest(t, "tcp")
}

func TestNetConnBigBufferedSendAfterClose(t *testing.T) {
	abstractConnBigBufferedSendAfterCloseTest(t, "tcp")
}

func TestNetConnCloseWhileReadError(t *testing.T) {
	conn2, done := connectionPair("tcp", func(conn1 Conn) {
		pkt := packet.NewPublishPacket()
		pkt.Message.Topic = "foo/bar/baz"
		buf := make([]byte, pkt.Len())
		pkt.Encode(buf)

		netConn := conn1.(*NetConn)
		_, err := netConn.UnderlyingConn().Write(buf[0:7]) // <- incomplete packet
		assert.NoError(t, err)

		err = netConn.UnderlyingConn().Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func TestNetConnCloseWhileDetectError(t *testing.T) {
	conn2, done := connectionPair("tcp", func(conn1 Conn) {
		pkt := packet.NewPublishPacket()
		pkt.Message.Topic = "foo/bar/baz"
		buf := make([]byte, pkt.Len())
		pkt.Encode(buf)

		netConn := conn1.(*NetConn)
		_, err := netConn.UnderlyingConn().Write(buf[0:1]) // <- too less for a detection
		assert.NoError(t, err)

		err = netConn.UnderlyingConn().Close()
		assert.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func TestNetConnReadTimeoutAfterDetect(t *testing.T) {
	conn2, done := connectionPair("tcp", func(conn1 Conn) {
		pkt := packet.NewPublishPacket()
		pkt.Message.Topic = "foo/bar/baz"
		buf := make([]byte, pkt.Len())
		pkt.Encode(buf)

		netConn := conn1.(*NetConn)
		_, err := netConn.UnderlyingConn().Write(buf[0 : len(buf)-1]) // <- not all of the bytes
		assert.NoError(t, err)
	})

	conn2.SetReadTimeout(10 * time.Millisecond)

	pkt, err := conn2.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	<-done
}

func BenchmarkNetConn(b *testing.B) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "foo/bar/baz"

	conn2, done := connectionPair("tcp", func(conn1 Conn) {
		for i := 0; i < b.N; i++ {
			err := conn1.Send(pkt)
			if err != nil {
				panic(err)
			}
		}
	})

	for i := 0; i < b.N; i++ {
		_, err := conn2.Receive()
		if err != nil {
			panic(err)
		}
	}

	b.SetBytes(int64(pkt.Len() * 2))

	<-done
}

func BenchmarkNetConnBuffered(b *testing.B) {
	pkt := packet.NewPublishPacket()
	pkt.Message.Topic = "foo/bar/baz"

	conn2, done := connectionPair("tcp", func(conn1 Conn) {
		for i := 0; i < b.N; i++ {
			err := conn1.BufferedSend(pkt)
			if err != nil {
				panic(err)
			}
		}
	})

	for i := 0; i < b.N; i++ {
		_, err := conn2.Receive()
		if err != nil {
			panic(err)
		}
	}

	b.SetBytes(int64(pkt.Len() * 2))

	<-done
}
