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

func TestConnectInterface(t *testing.T) {
	msg := NewConnectMessage()

	require.Equal(t, msg.Type(), CONNECT)
	require.NotNil(t, msg.String())
}

func TestConnectMessageDecode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,   // Protocol Level
		206, // Connect Flags
		0,   // Keep Alive MSB
		10,  // Keep Alive LSB
		0,   // Client ID MSB
		7,   // Client ID LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // Will Topic MSB
		4, // Will Topic LSB
		'w', 'i', 'l', 'l',
		0,  // Will Message MSB
		12, // Will Message LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		0, // Username ID MSB
		7, // Username ID LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0,  // Password ID MSB
		10, // Password ID LSB
		'v', 'e', 'r', 'y', 's', 'e', 'c', 'r', 'e', 't',
	}

	msg := NewConnectMessage()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, 10, int(msg.KeepAlive))
	require.Equal(t, "surgemq", string(msg.ClientId))
	require.Equal(t, "will", string(msg.WillTopic))
	require.Equal(t, "send me home", string(msg.WillPayload))
	require.Equal(t, "surgemq", string(msg.Username))
	require.Equal(t, "verysecret", string(msg.Password))
}

func TestConnectMessageDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60, // <- remaining length to high
		0,  // Protocol String MSB
		5,  // Protocol String LSB
		'M', 'Q', 'T', 'T',
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError2(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		5, // Protocol String LSB <- invalid size
		'M', 'Q', 'T', 'T',
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError3(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		// Protocol Level <- missing
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError4(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		8,
		0,                       // Protocol String MSB
		5,                       // Protocol String LSB
		'M', 'Q', 'T', 'T', 'X', // <- wrong protocol string
		4,
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError5(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		5, // Protocol Level <- wrong id
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError6(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		// Connect Flags <- missing
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError7(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		1, // Connect Flags <- reserved bit set to one
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError8(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,  // Protocol Level
		24, // Connect Flags <- invalid qos
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError9(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		8, // Connect Flags <- will flag set to zero but others not
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError10(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,  // Protocol Level
		64, // Connect Flags <- password flag set but not username
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError11(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		0, // Connect Flags
		0, // Keep Alive MSB <- missing
		// Keep Alive LSB <- missing
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError12(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		0, // Connect Flags
		0, // Keep Alive MSB
		1, // Keep Alive LSB
		0, // Client ID MSB
		2, // Client ID LSB <- wrong size
		'x',
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError13(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		0, // Connect Flags <- clean session false
		0, // Keep Alive MSB
		1, // Keep Alive LSB
		0, // Client ID MSB
		0, // Client ID LSB
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError14(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		6, // Connect Flags
		0, // Keep Alive MSB
		1, // Keep Alive LSB
		0, // Client ID MSB
		0, // Client ID LSB
		0, // Will Topic MSB
		1, // Will Topic LSB <- wrong size
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError15(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		6, // Connect Flags
		0, // Keep Alive MSB
		1, // Keep Alive LSB
		0, // Client ID MSB
		0, // Client ID LSB
		0, // Will Topic MSB
		0, // Will Topic LSB
		0, // Will Payload MSB
		1, // Will Payload LSB <- wrong size
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError16(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,   // Protocol Level
		194, // Connect Flags
		0,   // Keep Alive MSB
		1,   // Keep Alive LSB
		0,   // Client ID MSB
		0,   // Client ID LSB
		0,   // Username MSB
		1,   // Username LSB <- wrong size
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageDecodeError17(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,   // Protocol Level
		194, // Connect Flags
		0,   // Keep Alive MSB
		1,   // Keep Alive LSB
		0,   // Client ID MSB
		0,   // Client ID LSB
		0,   // Username MSB
		0,   // Username LSB
		0,   // Password MSB
		1,   // Password LSB <- wrong size
	}

	msg := NewConnectMessage()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestConnectMessageEncode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60,
		0, // Protocol String MSB (0)
		4, // Protocol String LSB (4)
		'M', 'Q', 'T', 'T',
		4,   // Protocol level 4
		206, // connect flags 11001110, will QoS = 01
		0,   // Keep Alive MSB (0)
		10,  // Keep Alive LSB (10)
		0,   // Client ID MSB (0)
		7,   // Client ID LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // Will Topic MSB (0)
		4, // Will Topic LSB (4)
		'w', 'i', 'l', 'l',
		0,  // Will Message MSB (0)
		12, // Will Message LSB (12)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		0, // Username ID MSB (0)
		7, // Username ID LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0,  // Password ID MSB (0)
		10, // Password ID LSB (10)
		'v', 'e', 'r', 'y', 's', 'e', 'c', 'r', 'e', 't',
	}

	msg := NewConnectMessage()
	msg.WillQoS = QosAtLeastOnce
	msg.Version = 4
	msg.CleanSession = true
	msg.ClientId = []byte("surgemq")
	msg.KeepAlive = 10
	msg.WillTopic = []byte("will")
	msg.WillPayload = []byte("send me home")
	msg.Username = []byte("surgemq")
	msg.Password = []byte("verysecret")

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestConnectEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60,
		0, // Protocol String MSB (0)
		4, // Protocol String LSB (4)
		'M', 'Q', 'T', 'T',
		4,   // Protocol level 4
		206, // connect flags 11001110, will QoS = 01
		0,   // Keep Alive MSB (0)
		10,  // Keep Alive LSB (10)
		0,   // Client ID MSB (0)
		7,   // Client ID LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // Will Topic MSB (0)
		4, // Will Topic LSB (4)
		'w', 'i', 'l', 'l',
		0,  // Will Message MSB (0)
		12, // Will Message LSB (12)
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
		0, // Username ID MSB (0)
		7, // Username ID LSB (7)
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0,  // Password ID MSB (0)
		10, // Password ID LSB (10)
		'v', 'e', 'r', 'y', 's', 'e', 'c', 'r', 'e', 't',
	}

	msg := NewConnectMessage()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)

	dst := make([]byte, msg.Len())
	n2, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, err := msg.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n3)
}

func BenchmarkConnectEncode(b *testing.B) {
	msg := NewConnectMessage()
	msg.WillQoS = QosAtLeastOnce
	msg.Version = 4
	msg.CleanSession = true
	msg.ClientId = []byte("i")
	msg.KeepAlive = 10
	msg.WillTopic = []byte("w")
	msg.WillPayload = []byte("m")
	msg.Username = []byte("u")
	msg.Password = []byte("p")

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkConnectDecode(b *testing.B) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		25,
		0, // Protocol String MSB (0)
		4, // Protocol String LSB (4)
		'M', 'Q', 'T', 'T',
		4,   // Protocol level 4
		206, // connect flags 11001110, will QoS = 01
		0,   // Keep Alive MSB (0)
		10,  // Keep Alive LSB (10)
		0,   // Client ID MSB (0)
		1,   // Client ID LSB (1)
		'i',
		0, // Will Topic MSB (0)
		1, // Will Topic LSB (1)
		'w',
		0, // Will Message MSB (0)
		1, // Will Message LSB (1)
		'm',
		0, // Username ID MSB (0)
		1, // Username ID LSB (1)
		'u',
		0, // Password ID MSB (0)
		1, // Password ID LSB (1)
		'p',
	}

	msg := NewConnectMessage()

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}
