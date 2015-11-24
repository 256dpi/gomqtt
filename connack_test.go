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

func TestConnackReturnCodes(t *testing.T) {
	require.Equal(t, ErrInvalidProtocolVersion.Error(), ConnackCode(1).Error(), "Incorrect ConnackCode error value.")
	require.Equal(t, ErrIdentifierRejected.Error(), ConnackCode(2).Error(), "Incorrect ConnackCode error value.")
	require.Equal(t, ErrServerUnavailable.Error(), ConnackCode(3).Error(), "Incorrect ConnackCode error value.")
	require.Equal(t, ErrBadUsernameOrPassword.Error(), ConnackCode(4).Error(), "Incorrect ConnackCode error value.")
	require.Equal(t, ErrNotAuthorized.Error(), ConnackCode(5).Error(), "Incorrect ConnackCode error value.")
}

func TestConnackMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	msg := NewConnackMessage()

	n, err := msg.Decode(msgBytes)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.False(t, msg.SessionPresent, "Error decoding session present flag.")
	require.Equal(t, ConnectionAccepted, msg.ReturnCode, "Error decoding return code.")
}

// testing wrong message length
func TestConnackMessageDecode2(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		3, // <- wrong size
		0, // session not present
		0, // connection accepted
	}

	msg := NewConnackMessage()

	_, err := msg.Decode(msgBytes)
	require.Error(t, err, "Error decoding message.")
}

// testing wrong message size
func TestConnackMessageDecode3(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
	}

	msg := NewConnackMessage()

	_, err := msg.Decode(msgBytes)
	require.Error(t, err, "Error decoding message.")
}

// testing wrong reserve bits
func TestConnackMessageDecode4(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		64, // <- wrong value
		0,  // connection accepted
	}

	msg := NewConnackMessage()

	_, err := msg.Decode(msgBytes)
	require.Error(t, err, "Error decoding message.")
}

// testing invalid return code
func TestConnackMessageDecode5(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		0,
		6, // <- wrong code
	}

	msg := NewConnackMessage()

	_, err := msg.Decode(msgBytes)
	require.Error(t, err, "Error decoding message.")
}

func TestConnackMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		1, // session present
		0, // connection accepted
	}

	msg := NewConnackMessage()
	msg.ReturnCode = ConnectionAccepted
	msg.SessionPresent = true

	dst := make([]byte, 10)
	n, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error encoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error encoding connack message.")
}

func TestConnackEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	msg := NewConnackMessage()
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

func BenchmarkConnackEncode(b *testing.B) {
	msg := NewConnackMessage()
	msg.ReturnCode = ConnectionAccepted
	msg.SessionPresent = true

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkConnackDecode(b *testing.B) {
	msgBytes := []byte{
		byte(CONNACK << 4),
		2,
		0, // session not present
		0, // connection accepted
	}

	msg := NewConnackMessage()

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}
