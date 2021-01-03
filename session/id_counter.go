package session

import (
	"sync/atomic"

	"github.com/256dpi/gomqtt/packet"
)

// An IDCounter continuously counts packet ids.
type IDCounter struct {
	next uint64
}

// NewIDCounter returns a new counter.
func NewIDCounter() *IDCounter {
	return NewIDCounterWithNext(1)
}

// NewIDCounterWithNext returns a new counter that will emit the specified id
// as the next id.
func NewIDCounterWithNext(next packet.ID) *IDCounter {
	return &IDCounter{
		next: uint64(next - 1),
	}
}

// NextID will return the next id.
func (c *IDCounter) NextID() packet.ID {
	// get next non zero id
	var id uint64
	for id == 0 {
		id = atomic.AddUint64(&c.next, 1) % (1 << 16)
	}

	return packet.ID(id)
}

// Reset will reset the counter.
func (c *IDCounter) Reset() {
	atomic.StoreUint64(&c.next, 0)
}
