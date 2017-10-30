package future

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	f := New()

	store := NewStore()
	assert.Equal(t, 0, len(store.All()))

	store.Put(1, f)
	assert.Equal(t, f, store.Get(1))
	assert.Equal(t, 1, len(store.All()))

	store.Delete(1)
	assert.Nil(t, store.Get(1))
	assert.Equal(t, 0, len(store.All()))
}

func TestStoreAwait(t *testing.T) {
	f := New()

	store := NewStore()
	assert.Equal(t, 0, len(store.All()))

	store.Put(1, f)
	assert.Equal(t, f, store.Get(1))
	assert.Equal(t, 1, len(store.All()))

	go func() {
		time.Sleep(1 * time.Millisecond)
		f.Complete()
		store.Delete(1)
	}()

	err := store.Await(10 * time.Millisecond)
	assert.NoError(t, err)
}

func TestStoreAwaitTimeout(t *testing.T) {
	f := New()

	store := NewStore()
	assert.Equal(t, 0, len(store.All()))

	store.Put(1, f)
	assert.Equal(t, f, store.Get(1))
	assert.Equal(t, 1, len(store.All()))

	err := store.Await(10 * time.Millisecond)
	assert.Equal(t, ErrTimeout, err)
}
