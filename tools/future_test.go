package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFutureCompleteBefore(t *testing.T) {
	f := NewFuture()
	f.Complete()
	assert.NoError(t, f.Wait(10*time.Millisecond))
}

func TestFutureCompleteAfter(t *testing.T) {
	done := make(chan struct{})

	f := NewFuture()

	go func() {
		assert.NoError(t, f.Wait(10*time.Millisecond))
		close(done)
	}()

	f.Complete()

	<-done
}

func TestFutureCancelBefore(t *testing.T) {
	f := NewFuture()
	f.Cancel()
	assert.Equal(t, ErrFutureCanceled, f.Wait(10*time.Millisecond))
}

func TestFutureCancelAfter(t *testing.T) {
	done := make(chan struct{})

	f := NewFuture()

	go func() {
		assert.Equal(t, ErrFutureCanceled, f.Wait(10*time.Millisecond))
		close(done)
	}()

	f.Cancel()

	<-done
}

func TestFutureTimeout(t *testing.T) {
	f := NewFuture()
	assert.Equal(t, ErrFutureTimeout, f.Wait(1*time.Millisecond))
}

func TestFutureBindBefore(t *testing.T) {
	done := make(chan struct{})

	f := NewFuture()
	f.Cancel()

	ff := NewFuture()
	go ff.Bind(f, nil)

	go func() {
		err := ff.Wait(10 * time.Millisecond)
		assert.Equal(t, ErrFutureCanceled, err)
		close(done)
	}()

	<-done
}

func TestFutureBindAfter(t *testing.T) {
	done := make(chan struct{})

	f := NewFuture()
	f.Cancel()

	ff := NewFuture()

	go func() {
		err := ff.Wait(10 * time.Millisecond)
		assert.Equal(t, ErrFutureCanceled, err)
		close(done)
	}()

	go ff.Bind(f, nil)

	<-done
}
