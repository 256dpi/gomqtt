package tools

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

	if q.head == q.tail && q.count > 0 {
		nodes := make([]*packet.Message, len(q.nodes)+q.size)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}

	q.nodes[q.tail] = msg
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a message from the queue in first to last order.
func (q *MessageQueue) Pop() *packet.Message {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.count == 0 {
		return nil
	}

	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--

	return node
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

	q.nodes = make([]*packet.Message, q.size)
	q.head = 0
	q.tail = 0
	q.count = 0
}
