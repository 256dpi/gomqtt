package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQOSCodes(t *testing.T) {
	assert.Equal(t, QOS(0), QOSAtMostOnce)
	assert.Equal(t, QOS(1), QOSAtLeastOnce)
	assert.Equal(t, QOS(2), QOSExactlyOnce)
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
		assert.Equal(t, d.name, m.String())
		assert.Equal(t, d.flags, m.defaultFlags())
	}
}

func TestDetect1(t *testing.T) {
	buf := []byte{0x10, 0x0}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 2, l)
	assert.Equal(t, 1, int(tt))
}

func TestDetect2(t *testing.T) {
	// not enough bytes
	buf := []byte{0x10, 0xff}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(tt))
}

func TestDetect3(t *testing.T) {
	buf := []byte{0x10, 0xff, 0x0}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 130, l)
	assert.Equal(t, 1, int(tt))
}

func TestDetect4(t *testing.T) {
	// not enough bytes
	buf := []byte{0x10, 0xff, 0xff}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(tt))
}

func TestDetect5(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x7f}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 268435460, l)
	assert.Equal(t, 1, int(tt))
}

func TestDetect6(t *testing.T) {
	buf := []byte{0x10}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(tt))
}

func TestDetect7(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x80}

	l, tt := DetectPacket(buf)
	assert.Equal(t, 0, l)
	assert.Equal(t, 0, int(tt))
}

func TestGetID(t *testing.T) {
	type detail struct {
		packet Generic
		ok     bool
	}

	details := map[Type]detail{
		CONNECT:     {NewConnect(), false},
		CONNACK:     {NewConnack(), false},
		PUBLISH:     {NewPublish(), true},
		PUBACK:      {NewPuback(), true},
		PUBREC:      {NewPubrec(), true},
		PUBREL:      {NewPubrel(), true},
		PUBCOMP:     {NewPubcomp(), true},
		SUBSCRIBE:   {NewSubscribe(), true},
		SUBACK:      {NewSuback(), true},
		UNSUBSCRIBE: {NewUnsubscribe(), true},
		UNSUBACK:    {NewUnsuback(), true},
		PINGREQ:     {NewPingreq(), false},
		PINGRESP:    {NewPingresp(), false},
		DISCONNECT:  {NewDisconnect(), false},
	}

	for _, d := range details {
		_, ok := GetID(d.packet)
		assert.Equal(t, d.ok, ok)
	}
}
