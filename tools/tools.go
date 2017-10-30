// Package tools implements utilities for building MQTT 3.1.1
// (http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) components.
package tools

import "sync"

// A Counter continuously counts packet ids.
type Counter struct {
	current uint16
	mutex   sync.Mutex
}

// NewCounter returns a new counter.
func NewCounter() *Counter {
	return &Counter{
		current: 1,
	}
}

// Next will return the next id.
func (c *Counter) Next() uint16 {
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
func (c *Counter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.current = 1
}
