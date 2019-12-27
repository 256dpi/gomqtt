package future

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFutureCompleteBefore(t *testing.T) {
	f := New()
	f.Complete(1)
	assert.NoError(t, f.Wait(10*time.Millisecond))
	assert.Equal(t, 1, f.Result())
}

func TestFutureCompleteAfter(t *testing.T) {
	done := make(chan struct{})

	f := New()

	go func() {
		assert.NoError(t, f.Wait(10*time.Millisecond))
		assert.Equal(t, 1, f.Result())
		close(done)
	}()

	f.Complete(1)

	<-done
}

func TestFutureCancelBefore(t *testing.T) {
	f := New()
	f.Cancel(1)
	assert.Equal(t, ErrCanceled, f.Wait(10*time.Millisecond))
	assert.Equal(t, 1, f.Result())
}

func TestFutureCancelAfter(t *testing.T) {
	done := make(chan struct{})

	f := New()

	go func() {
		assert.Equal(t, ErrCanceled, f.Wait(10*time.Millisecond))
		assert.Equal(t, 1, f.Result())
		close(done)
	}()

	f.Cancel(1)

	<-done
}

func TestFutureTimeout(t *testing.T) {
	f := New()
	assert.Equal(t, ErrTimeout, f.Wait(1*time.Millisecond))
}

func TestFutureBindBefore(t *testing.T) {
	done := make(chan struct{})

	f := New()
	f.Cancel(1)

	ff := New()
	ff.Bind(f)

	go func() {
		err := ff.Wait(10 * time.Millisecond)
		assert.Equal(t, ErrCanceled, err)
		assert.Equal(t, 1, ff.Result())
		close(done)
	}()

	<-done
}

func TestFutureBindAfter(t *testing.T) {
	done := make(chan struct{})

	f := New()
	f.Cancel(1)

	ff := New()

	go func() {
		err := ff.Wait(10 * time.Millisecond)
		assert.Equal(t, ErrCanceled, err)
		assert.Equal(t, 1, ff.Result())
		close(done)
	}()

	ff.Bind(f)

	<-done
}
