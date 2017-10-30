package client

import (
	"sync"
	"time"

	"github.com/256dpi/gomqtt/client/future"
	"github.com/256dpi/gomqtt/packet"
)

/* futureStore */

// a futureStore is used to store active Futures
type futureStore struct {
	sync.RWMutex

	protected bool
	store     map[packet.ID]*future.Future
}

// newFutureStore will create a new futureStore
func newFutureStore() *futureStore {
	return &futureStore{
		store: make(map[packet.ID]*future.Future),
	}
}

// put will save a GenericFuture to the store
func (s *futureStore) put(id packet.ID, future *future.Future) {
	s.Lock()
	defer s.Unlock()

	s.store[id] = future
}

// get will retrieve a GenericFuture from the store
func (s *futureStore) get(id packet.ID) *future.Future {
	s.RLock()
	defer s.RUnlock()

	return s.store[id]
}

// del will remove a GenericFuture from the store
func (s *futureStore) del(id packet.ID) {
	s.Lock()
	defer s.Unlock()

	delete(s.store, id)
}

// return a slice with all stored futures
func (s *futureStore) all() []*future.Future {
	s.RLock()
	defer s.RUnlock()

	all := make([]*future.Future, len(s.store))

	i := 0
	for _, savedFuture := range s.store {
		all[i] = savedFuture
		i++
	}

	return all
}

// set the protection attribute and if true prevents the store from being cleared
func (s *futureStore) protect(value bool) {
	s.Lock()
	defer s.Unlock()

	s.protected = value
}

// will cancel all stored futures and remove them if the store is unprotected
func (s *futureStore) clear() {
	s.Lock()
	defer s.Unlock()

	if s.protected {
		return
	}

	for _, savedFuture := range s.store {
		savedFuture.Cancel()
	}

	s.store = make(map[packet.ID]*future.Future)
}

// will wait until all futures have completed and removed or timeout is reached
func (s *futureStore) await(timeout time.Duration) error {
	stop := time.Now().Add(timeout)

	for {
		// get futures
		futures := s.all()

		// return if no futures are left
		if len(futures) == 0 {
			return nil
		}

		// wait for next future to complete
		err := futures[0].Wait(stop.Sub(time.Now()))
		if err != nil {
			return err
		}
	}
}

/* tracker */

// a tracker keeps track of keep alive intervals
type tracker struct {
	sync.Mutex

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
	t.Lock()
	defer t.Unlock()

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
	t.Lock()
	defer t.Unlock()

	return t.pings > 0
}
