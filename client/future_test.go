package client

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client/future"
	"github.com/stretchr/testify/assert"
)

func TestGenericFutureBind(t *testing.T) {
	f1 := newGenericFuture()
	f1.Cancel()

	f2 := newGenericFuture()
	go f2.Bind(f1)

	err := f2.Wait(10 * time.Millisecond)
	assert.Equal(t, future.ErrFutureCanceled, err)
}

func TestSubscribeFutureBind(t *testing.T) {
	f1 := newSubscribeFuture()
	f1.Cancel()

	f2 := newSubscribeFuture()
	go f2.Bind(f1)

	err := f2.Wait(10 * time.Millisecond)
	assert.Equal(t, future.ErrFutureCanceled, err)
}
