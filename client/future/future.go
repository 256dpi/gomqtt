package future

import (
	"errors"
	"time"
)

// ErrFutureTimeout is returned by Wait if the specified timeout is exceeded.
var ErrFutureTimeout = errors.New("future timeout")

// ErrFutureCanceled is returned by Wait if the future gets canceled while waiting.
var ErrFutureCanceled = errors.New("future canceled")

// A Future is a low-level future type that can be extended to transport
// custom information.
type Future struct {
	completeChannel chan struct{}
	cancelChannel   chan struct{}
}

// New will return a new Future.
func New() *Future {
	return &Future{
		completeChannel: make(chan struct{}),
		cancelChannel:   make(chan struct{}),
	}
}

// Bind will tie the current future to the specified future. If the bound to
// future is completed or canceled the current will as well.
func (f *Future) Bind(f2 *Future, transfer func()) {
	select {
	case <-f2.completeChannel:
		// call transfer function if available
		if transfer != nil {
			transfer()
		}

		close(f.completeChannel)
	case <-f2.cancelChannel:
		// call transfer function if available
		if transfer != nil {
			transfer()
		}

		close(f.cancelChannel)
	}
}

// Wait will wait the given amount of time and return whether the future has been
// completed, canceled or the request timed out.
func (f *Future) Wait(timeout time.Duration) error {
	select {
	case <-f.completeChannel:
		return nil
	case <-f.cancelChannel:
		return ErrFutureCanceled
	case <-time.After(timeout):
		return ErrFutureTimeout
	}
}

// Complete will complete the future.
func (f *Future) Complete() {
	// return if future has already been canceled
	select {
	case <-f.cancelChannel:
		return
	default:
	}

	close(f.completeChannel)
}

// Cancel will cancel the future.
func (f *Future) Cancel() {
	// return if future has already been completed
	select {
	case <-f.completeChannel:
		return
	default:
	}

	close(f.cancelChannel)
}
