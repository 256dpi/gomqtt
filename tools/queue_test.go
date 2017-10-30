package tools

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueue(t *testing.T) {
	msg := &packet.Message{}

	queue := NewQueue(2)
	assert.Equal(t, 0, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 1, queue.Len())

	msg1 := queue.Pop()
	assert.Equal(t, msg, msg1)
	assert.Equal(t, 0, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 1, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 2, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 2, queue.Len())

	messages := queue.All()
	assert.Equal(t, 0, queue.Len())
	assert.Equal(t, []*packet.Message{msg, msg}, messages)

	msg2 := queue.Pop()
	assert.Nil(t, msg2)
	assert.Equal(t, 0, queue.Len())
}
