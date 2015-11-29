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

package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnsubscribeInterface(t *testing.T) {
	msg := NewUnsubscribeMessage()
	msg.Topics = [][]byte{[]byte("hello")}

	require.Equal(t, msg.Type(), UNSUBSCRIBE)
	require.NotNil(t, msg.String())
}

func TestUnsubscribeMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		33,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	msg := NewUnsubscribeMessage()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, 3, len(msg.Topics))
	require.Equal(t, []byte("surgemq"), msg.Topics[0])
	require.Equal(t, []byte("/a/b/#/c"), msg.Topics[1])
	require.Equal(t, []byte("/a/b/#/cdd"), msg.Topics[2])
}

func TestUnsubscribeMessageDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		2,
		0, // packet ID MSB
		7, // packet ID LSB
		// empty topic list
	}

	msg := NewUnsubscribeMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestUnsubscribeMessageDecodeError2(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		6, // <- wrong remaining length
		0, // packet ID MSB
		7, // packet ID LSB
	}

	msg := NewUnsubscribeMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestUnsubscribeMessageDecodeError3(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		0,
		// missing packet id
	}

	msg := NewUnsubscribeMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestUnsubscribeMessageDecodeError4(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		11,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		9, // topic name LSB <- wrong size
		's', 'u', 'r', 'g', 'e', 'm', 'q',
	}

	msg := NewUnsubscribeMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestUnsubscribeMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		33,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	msg := NewUnsubscribeMessage()
	msg.PacketId = 7
	msg.Topics = [][]byte{
		[]byte("surgemq"),
		[]byte("/a/b/#/c"),
		[]byte("/a/b/#/cdd"),
	}

	dst := make([]byte, 100)
	n, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestUnsubscribeMessageEncodeError1(t *testing.T) {
	msg := NewUnsubscribeMessage()
	msg.PacketId = 7
	msg.Topics = [][]byte{
		[]byte("surgemq"),
	}

	dst := make([]byte, 1)
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestUnsubscribeMessageEncodeError2(t *testing.T) {
	msg := NewUnsubscribeMessage()
	msg.PacketId = 7
	msg.Topics = [][]byte{
		make([]byte, 65536),
	}

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 6, n)
}

// test to ensure encoding and decoding are the same
// decode, encode, and decode again
func TestUnsubscribeEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		33,
		0, // packet ID MSB
		7, // packet ID LSB
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // topic name MSB
		8, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c',
		0,  // topic name MSB
		10, // topic name LSB
		'/', 'a', '/', 'b', '/', '#', '/', 'c', 'd', 'd',
	}

	msg := NewUnsubscribeMessage()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)

	dst := make([]byte, 100)
	n2, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, err := msg.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n3)
}

func BenchmarkUnsubscribeEncode(b *testing.B) {
	msg := NewUnsubscribeMessage()
	msg.PacketId = 1
	msg.Topics = [][]byte{
		[]byte("t"),
	}

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkUnsubscribeDecode(b *testing.B) {
	msgBytes := []byte{
		byte(UNSUBSCRIBE<<4) | 2,
		5,
		0, // packet ID MSB
		1, // packet ID LSB
		0, // topic name MSB
		1, // topic name LSB
		't',
	}

	msg := NewUnsubscribeMessage()

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}
