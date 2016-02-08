// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFutureStore(t *testing.T) {
	future := &ConnectFuture{}

	store := newFutureStore()
	assert.Equal(t, 0, len(store.all()))

	store.put(1, future)
	assert.Equal(t, future, store.get(1))
	assert.Equal(t, 1, len(store.all()))

	store.del(1)
	assert.Nil(t, store.get(1))
	assert.Equal(t, 0, len(store.all()))
}

func TestFutureStoreAwait(t *testing.T) {
	future := &ConnectFuture{}
	future.initialize()

	store := newFutureStore()
	assert.Equal(t, 0, len(store.all()))

	store.put(1, future)
	assert.Equal(t, future, store.get(1))
	assert.Equal(t, 1, len(store.all()))

	go func() {
		time.Sleep(1 * time.Millisecond)
		future.complete()
		store.del(1)
	}()

	err := store.await(10 * time.Millisecond)
	assert.NoError(t, err)
}

func TestFutureStoreAwaitTimeout(t *testing.T) {
	future := &ConnectFuture{}
	future.initialize()

	store := newFutureStore()
	assert.Equal(t, 0, len(store.all()))

	store.put(1, future)
	assert.Equal(t, future, store.get(1))
	assert.Equal(t, 1, len(store.all()))

	err := store.await(10 * time.Millisecond)
	assert.Equal(t, ErrFutureTimeout, err)
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
