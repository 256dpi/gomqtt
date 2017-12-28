package client

import (
	"sync"
	"time"
)

// a tracker keeps track of keep alive intervals
type tracker struct {
	sync.RWMutex

	last    time.Time
	pings   uint8
	timeout time.Duration
}

// returns a new tracker
func newTracker(timeout time.Duration) *tracker {
	return &tracker{
		last:    time.Now(),
		timeout: timeout,
	}
}

// updates the tracker
func (t *tracker) reset() {
	t.Lock()
	defer t.Unlock()

	t.last = time.Now()
}

// returns the current time window
func (t *tracker) window() time.Duration {
	t.RLock()
	defer t.RUnlock()

	return t.timeout - time.Since(t.last)
}

// mark ping
func (t *tracker) ping() {
	t.Lock()
	defer t.Unlock()

	t.pings++
}

// mark pong
func (t *tracker) pong() {
	t.Lock()
	defer t.Unlock()

	t.pings--
}

// returns if pings are pending
func (t *tracker) pending() bool {
	t.RLock()
	defer t.RUnlock()

	return t.pings > 0
}
