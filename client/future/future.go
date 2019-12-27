// Package future implements a generic future handling system.
package future

import (
	"errors"
	"sync"
	"time"
)

// ErrTimeout is returned by Wait if the specified timeout is exceeded.
var ErrTimeout = errors.New("future timeout")

// ErrCanceled is returned by Wait if the future gets canceled while waiting.
var ErrCanceled = errors.New("future canceled")

// A Future is a low-level future type that can be extended to transport
// custom information.
type Future struct {
	result    interface{}
	completed chan struct{}
	cancelled chan struct{}
	done      bool
	mutex     sync.Mutex
}

// New will return a new Future.
func New() *Future {
	return &Future{
		completed: make(chan struct{}),
		cancelled: make(chan struct{}),
	}
}

// Bind will tie the current future to the specified future. If the bound to
// future is completed or canceled the current will as well. The result from the
// bound future is copied to the current on complete and cancel.
func (f *Future) Bind(f2 *Future) {
	go func() {
		err := f2.Wait(0)
		if err == ErrCanceled {
			f.Cancel(f2.Result())
		} else {
			f.Complete(f2.Result())
		}
	}()
}

// Wait will wait the given amount of time and return whether the future has been
// completed, canceled or the request timed out. If no time has been provided
// the wait will never timeout.
func (f *Future) Wait(timeout time.Duration) error {
	// prepare deadline
	var deadline <-chan time.Time
	if timeout > 0 {
		deadline = time.After(timeout)
	}

	// wait completion, cancellation or timeout
	select {
	case <-f.completed:
		return nil
	case <-f.cancelled:
		return ErrCanceled
	case <-deadline:
		return ErrTimeout
	}
}

// Complete will complete the future.
func (f *Future) Complete(result interface{}) bool {
	// acquire mutex
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// check flag
	if f.done {
		return false
	}

	// set result
	f.result = result

	// signal completion
	close(f.completed)

	// set flag
	f.done = true

	return true
}

// Cancel will cancel the future.
func (f *Future) Cancel(result interface{}) bool {
	// acquire mutex
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// check flag
	if f.done {
		return false
	}

	// set result
	f.result = result

	// signal cancellation
	close(f.cancelled)

	// set flag
	f.done = true

	return true
}

// Result will return the value provided when the future has been completed or
// cancelled.
func (f *Future) Result() interface{} {
	// acquire mutex
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.result
}
