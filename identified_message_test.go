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

var imType MessageType = 1

func TestIdentifiedMessageFields(t *testing.T) {
	msg := &PubackMessage{}
	msg.Type = imType

	msg.PacketId = 100

	require.Equal(t, 100, int(msg.PacketId))
}

func TestIdentifiedMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(imType << 4),
		2,
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	msg := &PubackMessage{}
	msg.Type = imType
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, imType, msg.Type, "Error decoding message.")
	require.Equal(t, 7, int(msg.PacketId), "Error decoding message.")
}

// test insufficient bytes
func TestIdentifiedMessageDecode2(t *testing.T) {
	msgBytes := []byte{
		byte(imType << 4),
		2,
		7, // packet ID LSB (7)
	}

	msg := &PubackMessage{}
	msg.Type = imType
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestIdentifiedMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(imType << 4),
		2,
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	msg := &PubackMessage{}
	msg.Type = imType
	msg.PacketId = 7

	dst := make([]byte, 10)
	n, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error decoding message.")
}

func TestIdentifiedMessageEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(imType << 4),
		2,
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	msg := &PubackMessage{}
	msg.Type = imType
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")

	dst := make([]byte, 100)
	n2, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n2, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n2], "Error decoding message.")

	n3, err := msg.Decode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n3, "Error decoding message.")
}

func BenchmarkIdentifiedMessageEncode(b *testing.B) {
	msg := &PubackMessage{}
	msg.Type = imType
	msg.PacketId = 1

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkIdentifiedMessageDecode(b *testing.B) {
	msgBytes := []byte{
		byte(imType << 4),
		2,
		0, // packet ID MSB (0)
		1, // packet ID LSB (7)
	}

	msg := &PubackMessage{}
	msg.Type = imType

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}
