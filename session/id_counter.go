package session

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
)

// An IDCounter continuously counts packet ids.
type IDCounter struct {
	current packet.ID
	mutex   sync.Mutex
}

// NewIDCounter returns a new counter.
func NewIDCounter() *IDCounter {
	return &IDCounter{
		current: 1,
	}
}

// NextID will return the next id.
func (c *IDCounter) NextID() packet.ID {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// cache current id
	id := c.current

	// increment id
	c.current++

	// increment again if current id is zero
	if c.current == 0 {
		c.current++
	}

	return id
}

// Reset will reset the counter.
func (c *IDCounter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.current = 1
}
