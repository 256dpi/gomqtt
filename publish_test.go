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

func TestPublishMessageDecode1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, 7, int(msg.PacketId), "Error decoding message.")
	require.Equal(t, "surgemq", string(msg.Topic), "Error deocding topic name.")
	require.Equal(t, []byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'}, msg.Payload, "Error deocding payload.")
}

// test insufficient bytes
func TestPublishMessageDecode2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		26,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

// test qos = 0 and no client id
func TestPublishMessageDecode3(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	_, err := msg.Decode(msgBytes)

	require.NoError(t, err, "Error decoding message.")
}

func TestPublishMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	msg.Topic = []byte("surgemq")
	msg.QoS = QosAtLeastOnce
	msg.PacketId = 7
	msg.Payload = []byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'}

	dst := make([]byte, 100)
	n, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error decoding message.")
}

// test empty topic name
func TestPublishMessageEncode2(t *testing.T) {
	msg := NewPublishMessage()
	msg.Topic = []byte("")
	msg.PacketId = 7
	msg.Payload = []byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'}

	dst := make([]byte, 100)
	_, err := msg.Encode(dst)
	require.Error(t, err)
}

// test encoding qos = 0 and no packet id
func TestPublishMessageEncode3(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()
	msg.Topic = []byte("surgemq")
	msg.QoS = QosAtMostOnce
	msg.Payload = []byte{'s', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e'}

	dst := make([]byte, 100)
	n, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error decoding message.")
}

// test large message
func TestPublishMessageEncode4(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		137,
		8,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
	}

	payload := make([]byte, 1024)
	msgBytes = append(msgBytes, payload...)

	msg := NewPublishMessage()
	msg.Topic = []byte("surgemq")
	msg.QoS = QosAtMostOnce
	msg.Payload = payload

	require.Equal(t, len(msgBytes), msg.Len())

	dst := make([]byte, 1100)
	n, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error decoding message.")
}

func TestPublishEqualDecodeEncode2(t *testing.T) {
	msgBytes := []byte{50, 18, 0, 9, 103, 114, 101, 101, 116, 105, 110, 103, 115, 0, 1, 72, 101, 108, 108, 111}

	msg := NewPublishMessage()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n, "Error decoding message.")

	dst := make([]byte, 100)
	n2, err := msg.Encode(dst)

	require.NoError(t, err, "Error decoding message.")
	require.Equal(t, len(msgBytes), n2, "Error decoding message.")
	require.Equal(t, msgBytes, dst[:n], "Error decoding message.")
}

func TestPublishEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB (0)
		7, // topic name LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB (0)
		7, // packet ID LSB (7)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishMessage()

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

func BenchmarkPublishEncode(b *testing.B) {
	msg := NewPublishMessage()
	msg.Topic = []byte("t")
	msg.QoS = QosAtLeastOnce
	msg.PacketId = 1
	msg.Payload = []byte("p")

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPublishDecode(b *testing.B) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		6,
		0, // topic name MSB (0)
		1, // topic name LSB (7)
		't',
		0, // packet ID MSB (0)
		1, // packet ID LSB (7)
		'p',
	}

	msg := NewPublishMessage()

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err);
		}
	}
}

