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
	require.Equal(t, "surgemq", string(msg.ClientID))
	require.Equal(t, "will", string(msg.WillTopic))
	require.Equal(t, "send me home", string(msg.WillPayload))
	require.Equal(t, "surgemq", string(msg.Username))
	require.Equal(t, "verysecret", string(msg.Password))
}

func TestConnectMessageDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60, // <- wrong remaining length
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

func TestConnectMessageEncode1(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,   // Protocol level 4
		204, // Connect Flags
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
	msg.WillQOS = QOSAtLeastOnce
	msg.CleanSession = false
	msg.ClientID = []byte("surgemq")
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

func TestConnectMessageEncode2(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		12,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,  // Protocol level 4
		2,  // Connect Flags
		0,  // Keep Alive MSB
		10, // Keep Alive LSB
		0,  // Client ID MSB
		0,  // Client ID LSB
	}

	msg := NewConnectMessage()
	msg.CleanSession = true
	msg.KeepAlive = 10

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestConnectMessageEncodeError1(t *testing.T) {
	msg := NewConnectMessage()

	dst := make([]byte, 4) // <- too small buffer
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestConnectMessageEncodeError2(t *testing.T) {
	msg := NewConnectMessage()
	msg.WillTopic = []byte("t")
	msg.WillQOS = 3 // <- wrong qos

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 9, n)
}

func TestConnectMessageEncodeError3(t *testing.T) {
	msg := NewConnectMessage()
	msg.ClientID = make([]byte, 65536) // <- too big

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 14, n)
}

func TestConnectMessageEncodeError4(t *testing.T) {
	msg := NewConnectMessage()
	msg.WillTopic = make([]byte, 65536) // <- too big

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 16, n)
}

func TestConnectMessageEncodeError5(t *testing.T) {
	msg := NewConnectMessage()
	msg.WillTopic = []byte("t")
	msg.WillPayload = make([]byte, 65536) // <- too big

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 19, n)
}

func TestConnectMessageEncodeError6(t *testing.T) {
	msg := NewConnectMessage()
	msg.Username = make([]byte, 65536) // <- too big

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 16, n)
}

func TestConnectMessageEncodeError7(t *testing.T) {
	msg := NewConnectMessage()
	msg.Password = []byte("p") // <- missing username

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 14, n)
}

func TestConnectMessageEncodeError8(t *testing.T) {
	msg := NewConnectMessage()
	msg.Username = []byte("u")
	msg.Password = make([]byte, 65536) // <- too big

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 19, n)
}

func TestConnectEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(CONNECT << 4),
		60,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,   // Protocol level 4
		238, // Connect Flags
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
	msg.WillQOS = QOSAtLeastOnce
	msg.CleanSession = true
	msg.ClientID = []byte("i")
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
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,   // Protocol level 4
		206, // Connect Flags
		0,   // Keep Alive MSB
		10,  // Keep Alive LSB
		0,   // Client ID MSB
		1,   // Client ID LSB
		'i',
		0, // Will Topic MSB
		1, // Will Topic LSB
		'w',
		0, // Will Message MSB
		1, // Will Message LSB
		'm',
		0, // Username ID MSB
		1, // Username ID LSB
		'u',
		0, // Password ID MSB
		1, // Password ID LSB
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
