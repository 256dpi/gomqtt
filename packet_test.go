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

	"github.com/stretchr/testify/assert"
)

func TestQOSCodes(t *testing.T) {
	assert.Equal(t, byte(0), QOSAtMostOnce)
	assert.Equal(t, byte(1), QOSAtLeastOnce)
	assert.Equal(t, byte(2), QOSExactlyOnce)
}

func TestFixedHeaderFlags(t *testing.T) {
	type detail struct {
		name  string
		flags byte
	}

	details := map[Type]detail{
		CONNECT:     {"Connect", 0},
		CONNACK:     {"Connack", 0},
		PUBLISH:     {"Publish", 0},
		PUBACK:      {"Puback", 0},
		PUBREC:      {"Pubrec", 0},
		PUBREL:      {"Pubrel", 2},
		PUBCOMP:     {"Pubcomp", 0},
		SUBSCRIBE:   {"Subscribe", 2},
		SUBACK:      {"Suback", 0},
		UNSUBSCRIBE: {"Unsubscribe", 2},
		UNSUBACK:    {"Unsuback", 0},
		PINGREQ:     {"Pingreq", 0},
		PINGRESP:    {"Pingresp", 0},
		DISCONNECT:  {"Disconnect", 0},
	}

	for m, d := range details {
		if m.String() != d.name {
			t.Errorf("Expected %s, got %s", d.name, m)
		}

		if m.defaultFlags() != d.flags {
			t.Errorf("Expected %d, got %d", d.flags, m.defaultFlags())
		}
	}
}

func TestDetect1(t *testing.T) {
	buf := []byte{0x10, 0x0}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 2, l)
	assert.Equal(t, 1, int(_t))
}

func TestDetect2(t *testing.T) {
	// not enough bytes
	buf := []byte{0x10, 0xff}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(_t))
}

func TestDetect3(t *testing.T) {
	buf := []byte{0x10, 0xff, 0x0}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 130, l)
	assert.Equal(t, 1, int(_t))
}

func TestDetect4(t *testing.T) {
	// not enough bytes
	buf := []byte{0x10, 0xff, 0xff}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(_t))
}

func TestDetect5(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x7f}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 268435460, l)
	assert.Equal(t, 1, int(_t))
}

func TestDetect6(t *testing.T) {
	buf := []byte{0x10}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(_t))
}

func TestDetect7(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x80}

	l, _t := DetectPacket(buf)

	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(_t))
}

func TestPacketID(t *testing.T) {
	type detail struct {
		packet Packet
		ok     bool
	}

	details := map[Type]detail{
		CONNECT:     {NewConnectPacket(), false},
		CONNACK:     {NewConnackPacket(), false},
		PUBLISH:     {NewPublishPacket(), true},
		PUBACK:      {NewPubackPacket(), true},
		PUBREC:      {NewPubrecPacket(), true},
		PUBREL:      {NewPubrelPacket(), true},
		PUBCOMP:     {NewPubcompPacket(), true},
		SUBSCRIBE:   {NewSubscribePacket(), true},
		SUBACK:      {NewSubackPacket(), true},
		UNSUBSCRIBE: {NewUnsubscribePacket(), true},
		UNSUBACK:    {NewUnsubackPacket(), true},
		PINGREQ:     {NewPingreqPacket(), false},
		PINGRESP:    {NewPingrespPacket(), false},
		DISCONNECT:  {NewDisconnectPacket(), false},
	}

	for _, d := range details {
		_, ok := PacketID(d.packet)
		assert.Equal(t, d.ok, ok)
	}
}

func TestFuzz(t *testing.T) {
	// too small buffer
	assert.Equal(t, 1, Fuzz([]byte{}))

	// wrong packet type
	b1 := []byte{0 << 4, 0x00}
	assert.Equal(t, 0, Fuzz(b1))

	// wrong packet format
	b2 := []byte{2 << 4, 0x02, 0x00, 0x06}
	assert.Equal(t, 0, Fuzz(b2))

	// right packet format
	b3 := []byte{2 << 4, 0x02, 0x00, 0x01}
	assert.Equal(t, 1, Fuzz(b3))
}
