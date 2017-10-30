package client

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client/future"
	"github.com/stretchr/testify/assert"
)

func TestFutureStore(t *testing.T) {
	f := future.New()

	store := newFutureStore()
	assert.Equal(t, 0, len(store.all()))

	store.put(1, f)
	assert.Equal(t, f, store.get(1))
	assert.Equal(t, 1, len(store.all()))

	store.del(1)
	assert.Nil(t, store.get(1))
	assert.Equal(t, 0, len(store.all()))
}

func TestFutureStoreAwait(t *testing.T) {
	f := future.New()

	store := newFutureStore()
	assert.Equal(t, 0, len(store.all()))

	store.put(1, f)
	assert.Equal(t, f, store.get(1))
	assert.Equal(t, 1, len(store.all()))

	go func() {
		time.Sleep(1 * time.Millisecond)
		f.Complete()
		store.del(1)
	}()

	err := store.await(10 * time.Millisecond)
	assert.NoError(t, err)
}

func TestFutureStoreAwaitTimeout(t *testing.T) {
	f := future.New()

	store := newFutureStore()
	assert.Equal(t, 0, len(store.all()))

	store.put(1, f)
	assert.Equal(t, f, store.get(1))
	assert.Equal(t, 1, len(store.all()))

	err := store.await(10 * time.Millisecond)
	assert.Equal(t, future.ErrTimeout, err)
}

func TestTracker(t *testing.T) {
	tracker := newTracker(10 * time.Millisecond)
	assert.False(t, tracker.pending())
	assert.True(t, tracker.window() > 0)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, tracker.window() <= 0)

	tracker.reset()
	assert.True(t, tracker.window() > 0)

	tracker.ping()
	assert.True(t, tracker.pending())

	tracker.pong()
	assert.False(t, tracker.pending())
}
