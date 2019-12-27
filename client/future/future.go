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
	Data *sync.Map

	completeChannel chan struct{}
	cancelChannel   chan struct{}
	completeOnce    sync.Once
	cancelOnce      sync.Once
}

// New will return a new Future.
func New() *Future {
	return &Future{
		Data:            new(sync.Map),
		completeChannel: make(chan struct{}),
		cancelChannel:   make(chan struct{}),
	}
}

// Bind will tie the current future to the specified future. If the bound to
// future is completed or canceled the current will as well. Data saved in the
// bound future is copied to the current on complete and cancel.
func (f *Future) Bind(f2 *Future) {
	go func(){
		select {
		case <-f2.completeChannel:
			f.Data = f2.Data
			f.Complete()
		case <-f2.cancelChannel:
			f.Data = f2.Data
			f.Cancel()
		}
	}()
}

// Wait will wait the given amount of time and return whether the future has been
// completed, canceled or the request timed out.
func (f *Future) Wait(timeout time.Duration) error {
	select {
	case <-f.completeChannel:
		return nil
	case <-f.cancelChannel:
		return ErrCanceled
	case <-time.After(timeout):
		return ErrTimeout
	}
}

// Complete will complete the future.
func (f *Future) Complete() {
	f.completeOnce.Do(func() {
		close(f.completeChannel)
	})
}

// Cancel will cancel the future.
func (f *Future) Cancel() {
	f.cancelOnce.Do(func() {
		close(f.cancelChannel)
	})
}
