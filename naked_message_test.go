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

func TestNakedMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedMessageDecode(msgBytes, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n)
}

func TestNakedMessageDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(DISCONNECT << 4),
		1, // <- wrong remaining length
		0,
	}

	n, err := nakedMessageDecode(msgBytes, DISCONNECT)

	require.Error(t, err)
	require.Equal(t, 2, n)
}

func TestNakedMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	dst := make([]byte, nakedMessageLen())
	n, err := nakedMessageEncode(dst, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestNakedMessageEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedMessageDecode(msgBytes, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n)

	dst := make([]byte, nakedMessageLen())
	n2, err := nakedMessageEncode(dst, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, err := nakedMessageDecode(dst, DISCONNECT)

	require.NoError(t, err)
	require.Equal(t, 2, n3)
}

func BenchmarkNakedMessageEncode(b *testing.B) {
	buf := make([]byte, nakedMessageLen())

	for i := 0; i < b.N; i++ {
		_, err := nakedMessageEncode(buf, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkNakedMessageDecode(b *testing.B) {
	msgBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	for i := 0; i < b.N; i++ {
		_, err := nakedMessageDecode(msgBytes, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func testNakedMessageImplementation(t *testing.T, mt MessageType) {
	msg, err := mt.New()
	require.NoError(t, err)
	require.Equal(t, mt, msg.Type())
	require.NotEmpty(t, msg.String())

	buf := make([]byte, msg.Len())
	n, err := msg.Encode(buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	n, err = msg.Decode(buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)
}

func TestDisconnectImplementation(t *testing.T) {
	testNakedMessageImplementation(t, DISCONNECT)
}

func TestPingreqImplementation(t *testing.T) {
	testNakedMessageImplementation(t, PINGREQ)
}

func TestPingrespImplementation(t *testing.T) {
	testNakedMessageImplementation(t, PINGRESP)
}
