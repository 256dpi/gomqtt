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

func TestQOSCodes(t *testing.T) {
	if QOSAtMostOnce != 0 || QOSAtLeastOnce != 1 || QOSExactlyOnce != 2 {
		t.Errorf("QOS codes invalid")
	}
}

func TestFixedHeaderFlags(t *testing.T) {
	type detail struct {
		name  string
		flags byte
	}

	details := map[Type]detail{
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

func TestDetect1(t *testing.T) {
	buf := []byte{0x10, 0x0}

	l, _t := DetectPacket(buf)

	require.Equal(t, 2, l)
	require.Equal(t, 1, int(_t))
}

// not enough bytes
func TestDetect2(t *testing.T) {
	buf := []byte{0x10, 0xff}

	l, _t := DetectPacket(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(_t))
}

func TestDetect3(t *testing.T) {
	buf := []byte{0x10, 0xff, 0x0}

	l, _t := DetectPacket(buf)

	require.Equal(t, 130, l)
	require.Equal(t, 1, int(_t))
}

// not enough bytes
func TestDetect4(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff}

	l, _t := DetectPacket(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(_t))
}

func TestDetect5(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x1}

	l, _t := DetectPacket(buf)

	require.Equal(t, 4194308, l)
	require.Equal(t, 1, int(_t))
}

func TestDetect6(t *testing.T) {
	buf := []byte{0x10}

	l, _t := DetectPacket(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(_t))
}

func TestFuzz(t *testing.T) {
	// too small buffer
	require.Equal(t, 1, Fuzz([]byte{}))

	// wrong packet type
	b1 := []byte{0 << 4, 0x00}
	require.Equal(t, 0, Fuzz(b1))

	// wrong packet format
	b2 := []byte{2 << 4, 0x02, 0x00, 0x06}
	require.Equal(t, 0, Fuzz(b2))

	// right packet format
	b3 := []byte{2 << 4, 0x02, 0x00, 0x01}
	require.Equal(t, 1, Fuzz(b3))
}
