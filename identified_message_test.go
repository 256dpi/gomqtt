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

func TestIdentifiedMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	n, pid, err := identifiedMessageDecode(msgBytes, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, 7, int(pid))
}

func TestIdentifiedMessageDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		1, // <- wrong remaining length
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	n, pid, err := identifiedMessageDecode(msgBytes, PUBACK)

	require.Error(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, 0, int(pid))
}

func TestIdentifiedMessageDecodeError2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		7, // packet ID LSB (7)
		// <- insufficient bytes
	}

	n, pid, err := identifiedMessageDecode(msgBytes, PUBACK)

	require.Error(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, 0, int(pid))
}

func TestIdentifiedMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	dst := make([]byte, identifiedMessageLen())
	n, err := identifiedMessageEncode(dst, 7, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestIdentifiedMessageEncodeError1(t *testing.T) {
	dst := make([]byte, 3) // <- insufficient buffer
	n, err := identifiedMessageEncode(dst, 7, PUBACK)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestIdentifiedMessageEncodeError2(t *testing.T) {
	dst := make([]byte, 4)
	n, err := identifiedMessageEncode(dst, 7, RESERVED) // <- wrong message type

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestIdentifiedMessageEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
	}

	msg := &PubackMessage{}
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, 4, n)

	dst := make([]byte, 100)
	n2, err := identifiedMessageEncode(dst, 7, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, pid, err := identifiedMessageDecode(msgBytes, PUBACK)

	require.NoError(t, err)
	require.Equal(t, 4, n3)
	require.Equal(t, 7, int(pid))
}

func BenchmarkIdentifiedMessageEncode(b *testing.B) {
	msg := &PubackMessage{}
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
		byte(PUBACK << 4),
		2,
		0, // packet ID MSB (0)
		1, // packet ID LSB (7)
	}

	msg := &PubackMessage{}

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}

func testIdentifiedMessageImplementation(t *testing.T, mt MessageType) {
	msg, err := mt.New()
	require.NoError(t, err)
	require.Equal(t, mt, msg.Type())
	require.NotEmpty(t, msg.String())

	buf := make([]byte, msg.Len())
	n, err := msg.Encode(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)

	n, err = msg.Decode(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
}

func TestPubackImplementation(t *testing.T) {
	testIdentifiedMessageImplementation(t, PUBACK)
}

func TestPubcompImplementation(t *testing.T) {
	testIdentifiedMessageImplementation(t, PUBCOMP)
}

func TestPubrecImplementation(t *testing.T) {
	testIdentifiedMessageImplementation(t, PUBREC)
}

func TestPubrelImplementation(t *testing.T) {
	testIdentifiedMessageImplementation(t, PUBREL)
}

func TestUnsubackImplementation(t *testing.T) {
	testIdentifiedMessageImplementation(t, UNSUBACK)
}
