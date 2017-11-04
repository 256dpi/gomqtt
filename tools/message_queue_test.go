package tools

import (
	"testing"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestMessageQueue(t *testing.T) {
	msg1 := &packet.Message{Topic: "m1"}
	msg2 := &packet.Message{Topic: "m2"}
	msg3 := &packet.Message{Topic: "m3"}

	queue := NewMessageQueue(2)
	assert.Equal(t, 0, queue.Len())

	queue.Push(msg1)
	queue.Push(msg2)
	assert.Equal(t, 2, queue.Len())

	msg := queue.Pop()
	assert.Equal(t, msg, msg1)
	assert.Equal(t, 1, queue.Len())

	queue.Push(msg3)
	assert.Equal(t, 2, queue.Len())

	msg = queue.Pop()
	assert.Equal(t, msg, msg2)
	assert.Equal(t, 1, queue.Len())

	msg = queue.Pop()
	assert.Equal(t, msg, msg3)
	assert.Equal(t, 0, queue.Len())

	msg = queue.Pop()
	assert.Nil(t, msg)
	assert.Equal(t, 0, queue.Len())
}
