package client

import (
	"sync"
	"time"
)

// Tracker a tracker keeps track of keep alive intervals
type Tracker struct {
	sync.RWMutex

	last    time.Time
	pings   uint8
	timeout time.Duration
}

// NewTracker returns a new tracker
func NewTracker(timeout time.Duration) *Tracker {
	return &Tracker{
		last:    time.Now(),
		timeout: timeout,
	}
}

// Reset updates the tracker
func (t *Tracker) Reset() {
	t.Lock()
	defer t.Unlock()

	t.last = time.Now()
}

// Window returns the current time window
func (t *Tracker) Window() time.Duration {
	t.RLock()
	defer t.RUnlock()

	return t.timeout - time.Since(t.last)
}

// Ping mark ping
func (t *Tracker) Ping() {
	t.Lock()
	defer t.Unlock()

	t.pings++
}

// Pong mark pong
func (t *Tracker) Pong() {
	t.Lock()
	defer t.Unlock()

	t.pings--
}

// Pending returns if pings are pending
func (t *Tracker) Pending() bool {
	t.RLock()
	defer t.RUnlock()

	return t.pings > 0
}
