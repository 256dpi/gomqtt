package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageString(t *testing.T) {
	msg := &Message{
		Topic:   "w",
		Payload: []byte("m"),
		QOS:     QOSAtLeastOnce,
	}

	assert.Equal(t, "<Message Topic=\"w\" QOS=1 Retain=false Payload=6d>", msg.String())
}

func TestMessageCopy(t *testing.T) {
	msg1 := &Message{
		Topic:   "w",
		Payload: []byte("m"),
		QOS:     QOSAtLeastOnce,
	}

	msg2 := msg1.Copy()
	assert.Equal(t, msg1.String(), msg2.String())

	msg1.Retain = true
	assert.False(t, msg2.Retain)
}
