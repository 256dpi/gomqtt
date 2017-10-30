package session

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
)

// NewMessageQueue returns a new MessageQueue. If maxSize is greater than zero the
// queue will not grow more than the defined size.
func NewMessageQueue(maxSize int) *MessageQueue {
	return &MessageQueue{
		maxSize: maxSize,
	}
}

// MessageQueue is a basic FIFO queue for messages.
type MessageQueue struct {
	maxSize int

	list  []*packet.Message
	mutex sync.Mutex
}

// Push adds a message to the queue.
func (q *MessageQueue) Push(msg *packet.Message) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.list) == q.maxSize {
		q.pop()
	}

	q.list = append(q.list, msg)
}

// Pop removes and returns a message from the queue in first to last order.
func (q *MessageQueue) Pop() *packet.Message {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.list) == 0 {
		return nil
	}

	return q.pop()
}

func (q *MessageQueue) pop() *packet.Message {
	x := len(q.list) - 1
	msg := q.list[x]
	q.list = q.list[:x]
	return msg
}

// Len returns the length of the queue.
func (q *MessageQueue) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.list)
}

// All returns and removes all messages from the queue.
func (q *MessageQueue) All() []*packet.Message {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	cache := q.list
	q.list = nil
	return cache
}
