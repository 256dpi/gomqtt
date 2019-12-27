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
	f := New()

	time.AfterFunc(time.Millisecond, func() {
		f.Complete(1)
	})

	assert.NoError(t, f.Wait(10*time.Millisecond))
	assert.Equal(t, 1, f.Result())
}

func TestFutureCancelBefore(t *testing.T) {
	f := New()
	f.Cancel(1)
	assert.Equal(t, ErrCanceled, f.Wait(10*time.Millisecond))
	assert.Equal(t, 1, f.Result())
}

func TestFutureCancelAfter(t *testing.T) {
	f := New()

	time.AfterFunc(time.Millisecond, func() {
		f.Cancel(1)
	})

	assert.Equal(t, ErrCanceled, f.Wait(10*time.Millisecond))
	assert.Equal(t, 1, f.Result())
}

func TestFutureTimeout(t *testing.T) {
	f := New()
	assert.Equal(t, ErrTimeout, f.Wait(1*time.Millisecond))
}

func TestFutureBindBefore(t *testing.T) {
	f := New()
	f.Cancel(1)

	ff := New()
	ff.Bind(f)

	err := ff.Wait(10 * time.Millisecond)
	assert.Equal(t, ErrCanceled, err)
	assert.Equal(t, 1, ff.Result())
}

func TestFutureBindAfter(t *testing.T) {
	f := New()

	time.AfterFunc(time.Millisecond, func() {
		f.Cancel(1)
	})

	ff := New()
	ff.Bind(f)

	err := ff.Wait(10 * time.Millisecond)
	assert.Equal(t, ErrCanceled, err)
	assert.Equal(t, 1, ff.Result())
}
