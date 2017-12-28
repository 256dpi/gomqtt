package broker

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
)

// MessageQueue is a basic FIFO queue for messages.
type MessageQueue struct {
	size int

	nodes []*packet.Message
	head  int
	tail  int
	count int

	mutex sync.RWMutex
}

// NewMessageQueue returns a new MessageQueue. If size is greater than zero the
// queue will not grow more than the defined size.
func NewMessageQueue(size int) *MessageQueue {
	return &MessageQueue{
		size:  size,
		nodes: make([]*packet.Message, size),
	}
}

// Push adds a message to the queue.
func (q *MessageQueue) Push(msg *packet.Message) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// remove item if full
	if q.count == q.size {
		q.Pop()
	}

	// add item
	q.nodes[q.head] = msg
	q.count++
	q.head = q.wrap(q.head + 1)
}

// Pop removes and returns a message from the queue in first to last order.
func (q *MessageQueue) Pop() *packet.Message {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.count == 0 {
		return nil
	}

	// remove item
	node := q.nodes[q.tail]
	q.nodes[q.tail] = nil
	q.count--
	q.tail = q.wrap(q.tail + 1)

	return node
}

// Range will call range with the contents of the queue. If fn returns false the
// operation is stopped immediately.
func (q *MessageQueue) Range(fn func(*packet.Message) bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	for i := 0; i < q.count; i++ {
		if !fn(q.nodes[q.wrap(q.head+i)]) {
			return
		}
	}
}

// Len returns the length of the queue.
func (q *MessageQueue) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.count
}

// Reset returns and removes all messages from the queue.
func (q *MessageQueue) Reset() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// reset state
	q.nodes = make([]*packet.Message, q.size)
	q.head = 0
	q.tail = 0
	q.count = 0
}

func (q *MessageQueue) wrap(i int) int {
	if i >= q.size {
		return i - q.size
	}

	return i
}
