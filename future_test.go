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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAbstractFuture(t *testing.T) {
	done := make(chan struct{})

	f := &abstractFuture{}
	f.initialize()

	go func() {
		assert.NoError(t, f.Wait())
		close(done)
	}()

	f.complete()

	<-done
}

func TestAbstractFutureCancel(t *testing.T) {
	done := make(chan struct{})

	f := &abstractFuture{}
	f.initialize()

	go func() {
		assert.Equal(t, ErrFutureCanceled, f.Wait())
		close(done)
	}()

	f.cancel()

	<-done
}

func TestAbstractFutureTimeout(t *testing.T) {
	done := make(chan struct{})

	f := &abstractFuture{}
	f.initialize()

	go func() {
		assert.NoError(t, f.Wait(10*time.Millisecond))
		close(done)
	}()

	f.complete()

	<-done
}

func TestAbstractFutureCancelTimeout(t *testing.T) {
	done := make(chan struct{})

	f := &abstractFuture{}
	f.initialize()

	go func() {
		assert.Equal(t, ErrFutureCanceled, f.Wait(10*time.Millisecond))
		close(done)
	}()

	f.cancel()

	<-done
}

func TestAbstractFutureTimeoutExceeded(t *testing.T) {
	done := make(chan struct{})

	f := &abstractFuture{}
	f.initialize()

	go func() {
		assert.Equal(t, ErrFutureTimeout, f.Wait(1*time.Millisecond))
		close(done)
	}()

	<-time.After(10 * time.Millisecond)

	f.complete()

	<-done
}

func TestAbstractFutureCall(t *testing.T) {
	done := make(chan struct{})

	f := &abstractFuture{}
	f.initialize()

	f.Call(func(err error) {
		assert.NoError(t, err)
		close(done)
	})

	f.complete()

	<-done
}

func TestPublishFutureBind(t *testing.T) {
	done := make(chan struct{})

	f := &PublishFuture{}
	f.initialize()

	ff := &PublishFuture{}
	ff.initialize()

	ff.bind(f)

	ff.Call(func(err error) {
		assert.Equal(t, ErrFutureCanceled, err)
		close(done)
	})

	f.cancel()

	<-done
}

func TestSubscribeFutureBind(t *testing.T) {
	done := make(chan struct{})

	f := &SubscribeFuture{}
	f.initialize()

	ff := &SubscribeFuture{}
	ff.initialize()

	ff.bind(f)

	ff.Call(func(err error) {
		assert.Equal(t, ErrFutureCanceled, err)
		close(done)
	})

	f.cancel()

	<-done
}

func TestUnsubscribeFutureBind(t *testing.T) {
	done := make(chan struct{})

	f := &UnsubscribeFuture{}
	f.initialize()

	ff := &UnsubscribeFuture{}
	ff.initialize()

	ff.bind(f)

	ff.Call(func(err error) {
		assert.Equal(t, ErrFutureCanceled, err)
		close(done)
	})

	f.cancel()

	<-done
}
