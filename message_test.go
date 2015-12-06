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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQosCodes(t *testing.T) {
	if QosAtMostOnce != 0 || QosAtLeastOnce != 1 || QosExactlyOnce != 2 {
		t.Errorf("QOS codes invalid")
	}
}

func TestFixedHeaderFlags(t *testing.T) {
	type detail struct {
		name  string
		flags byte
	}

	details := map[MessageType]detail{
		RESERVED:    {"RESERVED", 0},
		CONNECT:     {"CONNECT", 0},
		CONNACK:     {"CONNACK", 0},
		PUBLISH:     {"PUBLISH", 0},
		PUBACK:      {"PUBACK", 0},
		PUBREC:      {"PUBREC", 0},
		PUBREL:      {"PUBREL", 2},
		PUBCOMP:     {"PUBCOMP", 0},
		SUBSCRIBE:   {"SUBSCRIBE", 2},
		SUBACK:      {"SUBACK", 0},
		UNSUBSCRIBE: {"UNSUBSCRIBE", 2},
		UNSUBACK:    {"UNSUBACK", 0},
		PINGREQ:     {"PINGREQ", 0},
		PINGRESP:    {"PINGRESP", 0},
		DISCONNECT:  {"DISCONNECT", 0},
		RESERVED2:   {"RESERVED2", 0},
	}

	for m, d := range details {
		if m.String() != d.name {
			t.Errorf("Name mismatch. Expecting %s, got %s", d.name, m)
		}

		if m.defaultFlags() != d.flags {
			t.Errorf("Flag mismatch for %s. Expecting %d, got %d", m, d.flags, m.defaultFlags())
		}
	}
}

func ExampleReadme() {
	// Create new message.
	msg1 := NewConnectMessage()
	msg1.Username = []byte("gomqtt")
	msg1.Password = []byte("amazing!")

	// Allocate buffer.
	buf := make([]byte, msg1.Len())

	// Encode the message.
	if _, err := msg1.Encode(buf); err != nil {
		// there was an error while encoding
		panic(err)
	}

	// Send buffer off the wire.

	// ---

	// Get buffer from the wire.

	// Detect message.
	l, mt := DetectMessage(buf)

	// Check length
	if l == 0 {
		// buffer not complete yet
		return
	}

	// Create message.
	msg2, err := mt.New()
	if err != nil {
		// message type is invalid
		panic(err)
	}

	// Decode message.
	_, err = msg2.Decode(buf)
	if err != nil {
		// there was an error while decoding
		panic(err)
	}

	switch msg2.Type() {
	case CONNECT:
		c := msg2.(*ConnectMessage)
		fmt.Println(string(c.Username))
		fmt.Println(string(c.Password))
	}

	// Output:
	// gomqtt
	// amazing!
}

func TestDetect1(t *testing.T) {
	buf := []byte{0x10, 0x0}

	l, mt := DetectMessage(buf)

	require.Equal(t, 2, l)
	require.Equal(t, 1, int(mt))
}

// not enough bytes
func TestDetect2(t *testing.T) {
	buf := []byte{0x10, 0xff}

	l, mt := DetectMessage(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(mt))
}

func TestDetect3(t *testing.T) {
	buf := []byte{0x10, 0xff, 0x0}

	l, mt := DetectMessage(buf)

	require.Equal(t, 130, l)
	require.Equal(t, 1, int(mt))
}

// not enough bytes
func TestDetect4(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff}

	l, mt := DetectMessage(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(mt))
}

func TestDetect5(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x1}

	l, mt := DetectMessage(buf)

	require.Equal(t, 4194308, l)
	require.Equal(t, 1, int(mt))
}

func TestDetect6(t *testing.T) {
	buf := []byte{0x10}

	l, mt := DetectMessage(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(mt))
}

func TestFuzz(t *testing.T) {
	// too small buffer
	require.Equal(t, 1, Fuzz([]byte{}))

	// wrong message type
	b1 := []byte{0 << 4, 0x00}
	require.Equal(t, 0, Fuzz(b1))

	// wrong message format
	b2 := []byte{2 << 4, 0x02, 0x00, 0x06}
	require.Equal(t, 0, Fuzz(b2))

	// right message format
	b3 := []byte{2 << 4, 0x02, 0x00, 0x01}
	require.Equal(t, 1, Fuzz(b3))
}
