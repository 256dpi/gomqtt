package future

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFutureCompleteBefore(t *testing.T) {
	f := New()
	f.Complete()
	assert.NoError(t, f.Wait(10*time.Millisecond))
}

func TestFutureCompleteAfter(t *testing.T) {
	done := make(chan struct{})

	f := New()

	go func() {
		assert.NoError(t, f.Wait(10*time.Millisecond))
		close(done)
	}()

	f.Complete()

	<-done
}

func TestFutureCancelBefore(t *testing.T) {
	f := New()
	f.Cancel()
	assert.Equal(t, ErrCanceled, f.Wait(10*time.Millisecond))
}

func TestFutureCancelAfter(t *testing.T) {
	done := make(chan struct{})

	f := New()

	go func() {
		assert.Equal(t, ErrCanceled, f.Wait(10*time.Millisecond))
		close(done)
	}()

	f.Cancel()

	<-done
}

func TestFutureTimeout(t *testing.T) {
	f := New()
	assert.Equal(t, ErrTimeout, f.Wait(1*time.Millisecond))
}

func TestFutureBindBefore(t *testing.T) {
	done := make(chan struct{})

	f := New()
	f.Data.Store("foo", "bar")
	f.Cancel()

	ff := New()
	ff.Bind(f)

	go func() {
		err := ff.Wait(10 * time.Millisecond)
		assert.Equal(t, ErrCanceled, err)
		val, _ := ff.Data.Load("foo")
		assert.Equal(t, "bar", val)
		close(done)
	}()

	<-done
}

func TestFutureBindAfter(t *testing.T) {
	done := make(chan struct{})

	f := New()
	f.Data.Store("foo", "bar")
	f.Cancel()

	ff := New()

	go func() {
		err := ff.Wait(10 * time.Millisecond)
		assert.Equal(t, ErrCanceled, err)
		val, _ := ff.Data.Load("foo")
		assert.Equal(t, "bar", val)
		close(done)
	}()

	ff.Bind(f)

	<-done
}
