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

package packet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectInterface(t *testing.T) {
	pkt := NewConnectPacket()

	require.Equal(t, pkt.Type(), CONNECT)
	require.NotNil(t, pkt.String())
}

func TestConnectPacketDecode(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, 10, int(pkt.KeepAlive))
	require.Equal(t, "surgemq", string(pkt.ClientID))
	require.Equal(t, "will", string(pkt.WillTopic))
	require.Equal(t, "send me home", string(pkt.WillPayload))
	require.Equal(t, "surgemq", string(pkt.Username))
	require.Equal(t, "verysecret", string(pkt.Password))
}

func TestConnectPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		60, // <- wrong remaining length
		0,  // Protocol String MSB
		5,  // Protocol String LSB
		'M', 'Q', 'T', 'T',
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError2(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		5, // Protocol String LSB <- invalid size
		'M', 'Q', 'T', 'T',
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError3(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		6,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		// Protocol Level <- missing
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError4(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		8,
		0,                       // Protocol String MSB
		5,                       // Protocol String LSB
		'M', 'Q', 'T', 'T', 'X', // <- wrong protocol string
		4,
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError5(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		5, // Protocol Level <- wrong id
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError6(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		// Connect Flags <- missing
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError7(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		1, // Connect Flags <- reserved bit set to one
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError8(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,  // Protocol Level
		24, // Connect Flags <- invalid qos
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError9(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4, // Protocol Level
		8, // Connect Flags <- will flag set to zero but others not
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError10(t *testing.T) {
	pktBytes := []byte{
		byte(CONNECT << 4),
		7,
		0, // Protocol String MSB
		4, // Protocol String LSB
		'M', 'Q', 'T', 'T',
		4,  // Protocol Level
		64, // Connect Flags <- password flag set but not username
	}

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError11(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError12(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError13(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError14(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError15(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError16(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketDecodeError17(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	_, err := pkt.Decode(pktBytes)

	require.Error(t, err)
}

func TestConnectPacketEncode1(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	pkt.WillQOS = QOSAtLeastOnce
	pkt.CleanSession = false
	pkt.ClientID = []byte("surgemq")
	pkt.KeepAlive = 10
	pkt.WillTopic = []byte("will")
	pkt.WillPayload = []byte("send me home")
	pkt.Username = []byte("surgemq")
	pkt.Password = []byte("verysecret")

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, pktBytes, dst[:n])
}

func TestConnectPacketEncode2(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	pkt.CleanSession = true
	pkt.KeepAlive = 10

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)
	require.Equal(t, pktBytes, dst[:n])
}

func TestConnectPacketEncodeError1(t *testing.T) {
	pkt := NewConnectPacket()

	dst := make([]byte, 4) // <- too small buffer
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 0, n)
}

func TestConnectPacketEncodeError2(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.WillTopic = []byte("t")
	pkt.WillQOS = 3 // <- wrong qos

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 9, n)
}

func TestConnectPacketEncodeError3(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.ClientID = make([]byte, 65536) // <- too big

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 14, n)
}

func TestConnectPacketEncodeError4(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.WillTopic = make([]byte, 65536) // <- too big

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 16, n)
}

func TestConnectPacketEncodeError5(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.WillTopic = []byte("t")
	pkt.WillPayload = make([]byte, 65536) // <- too big

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 19, n)
}

func TestConnectPacketEncodeError6(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.Username = make([]byte, 65536) // <- too big

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 16, n)
}

func TestConnectPacketEncodeError7(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.Password = []byte("p") // <- missing username

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 14, n)
}

func TestConnectPacketEncodeError8(t *testing.T) {
	pkt := NewConnectPacket()
	pkt.Username = []byte("u")
	pkt.Password = make([]byte, 65536) // <- too big

	dst := make([]byte, pkt.Len())
	n, err := pkt.Encode(dst)

	require.Error(t, err)
	require.Equal(t, 19, n)
}

func TestConnectEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()
	n, err := pkt.Decode(pktBytes)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n)

	dst := make([]byte, pkt.Len())
	n2, err := pkt.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n2)
	require.Equal(t, pktBytes, dst[:n2])

	n3, err := pkt.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(pktBytes), n3)
}

func BenchmarkConnectEncode(b *testing.B) {
	pkt := NewConnectPacket()
	pkt.WillQOS = QOSAtLeastOnce
	pkt.CleanSession = true
	pkt.ClientID = []byte("i")
	pkt.KeepAlive = 10
	pkt.WillTopic = []byte("w")
	pkt.WillPayload = []byte("m")
	pkt.Username = []byte("u")
	pkt.Password = []byte("p")

	buf := make([]byte, pkt.Len())

	for i := 0; i < b.N; i++ {
		_, err := pkt.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkConnectDecode(b *testing.B) {
	pktBytes := []byte{
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

	pkt := NewConnectPacket()

	for i := 0; i < b.N; i++ {
		_, err := pkt.Decode(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}
