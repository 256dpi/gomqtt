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
	complete  sync.Once
	cancel    sync.Once
}

// New will return a new Future.
func New() *Future {
	return &Future{
		completed: make(chan struct{}),
		cancelled: make(chan struct{}),
	}
}

// Bind will tie the current future to the specified future. If the bound to
// future is completed or canceled the current will as well. Data saved in the
// bound future is copied to the current on complete and cancel.
func (f *Future) Bind(f2 *Future) {
	go func() {
		select {
		case <-f2.completed:
			f.Complete(f2.Result())
		case <-f2.cancelled:
			f.Cancel(f2.Result())
		}
	}()
}

// Wait will wait the given amount of time and return whether the future has been
// completed, canceled or the request timed out.
func (f *Future) Wait(timeout time.Duration) error {
	select {
	case <-f.completed:
		return nil
	case <-f.cancelled:
		return ErrCanceled
	case <-time.After(timeout):
		return ErrTimeout
	}
}

// Complete will complete the future.
func (f *Future) Complete(result interface{}) {
	f.complete.Do(func() {
		f.result = result
		close(f.completed)
	})
}

// Cancel will cancel the future.
func (f *Future) Cancel(result interface{}) {
	f.cancel.Do(func() {
		f.result = result
		close(f.cancelled)
	})
}

// Result will return the value provided when the future has been completed or
// cancelled.
func (f *Future) Result() interface{} {
	return f.result
}
