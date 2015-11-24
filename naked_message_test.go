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

var nmType MessageType = 1

func TestNakedMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(nmType << 4),
		0,
	}

	n, err := nakedMessageDecode(msgBytes, nmType)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
}

func TestNakedMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(nmType << 4),
		0,
	}

	dst := make([]byte, 10)
	n, err := nakedMessageEncode(dst, nmType)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error decoding message.")
}

func TestNakedMessageEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(nmType << 4),
		0,
	}

	n, err := nakedMessageDecode(msgBytes, nmType)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")

	dst := make([]byte, 100)
	n2, err := nakedMessageEncode(dst, nmType)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n2, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n2], "Error decoding message.")

	n3, err := nakedMessageDecode(dst, nmType)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n3, "Error decoding message.")
}

func BenchmarkNakedMessageEncode(b *testing.B) {
	buf := make([]byte, nakedMessageLen())

	for i := 0; i < b.N; i++ {
		_, err := nakedMessageEncode(buf, nmType)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkNakedMessageDecode(b *testing.B) {
	msgBytes := []byte{
		byte(nmType << 4),
		0,
	}

	for i := 0; i < b.N; i++ {
		_, err := nakedMessageDecode(msgBytes, nmType)
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
	_, err = msg.Encode(buf)
	require.NoError(t, err)

	_, err = msg.Decode(buf)
	require.NoError(t, err)
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
