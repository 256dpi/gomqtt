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
	"errors"
	"time"

	"github.com/gomqtt/packet"
)

// ErrFutureTimeout is returned by Wait if the specified timeout is exceeded.
var ErrFutureTimeout = errors.New("future timeout")

// ErrFutureCanceled is returned by Wait if the future gets canceled while waiting.
var ErrFutureCanceled = errors.New("future canceled")

// A Future represents information that might become available in the future.
type Future interface {
	// Wait will block until the future is completed or canceled. It will return
	// ErrCanceled if the future gets canceled. If a timeout is specified it
	// might return a ErrTimeoutExceeded.
	//
	// Note: Wait will not return an Client related errors.
	Wait(timeout ...time.Duration) error

	// Call calls the supplied callback in a separate goroutine with the error
	// returned by Wait.
	Call(func(err error), ...time.Duration)
}

type abstractFuture struct {
	completeChannel chan struct{}
	cancelChannel   chan struct{}
}

func (f *abstractFuture) initialize() {
	f.completeChannel = make(chan struct{})
	f.cancelChannel = make(chan struct{})
}

func (f *abstractFuture) Wait(timeout ...time.Duration) error {
	if len(timeout) > 0 {
		select {
		case <-f.completeChannel:
			return nil
		case <-f.cancelChannel:
			return ErrFutureCanceled
		case <-time.After(timeout[0]):
			return ErrFutureTimeout
		}
	}

	select {
	case <-f.completeChannel:
		return nil
	case <-f.cancelChannel:
		return ErrFutureCanceled
	}
}

func (f *abstractFuture) Call(callback func(error), timeout ...time.Duration) {
	go func() {
		callback(f.Wait(timeout...))
	}()
}

func (f *abstractFuture) complete() {
	close(f.completeChannel)
}

func (f *abstractFuture) cancel() {
	close(f.cancelChannel)
}

// The ConnectFuture is returned by the Client on Connect.
type ConnectFuture struct {
	abstractFuture

	SessionPresent bool
	ReturnCode     packet.ConnackCode
}

// The PublishFuture is returned by the Client on Publish.
type PublishFuture struct {
	abstractFuture
}

func (f *PublishFuture) bind(future Future) {
	future.Call(func(err error) {
		if err != nil {
			f.cancel()
			return
		}

		f.complete()
	})
}

// The SubscribeFuture is returned by the Client on Subscribe.
type SubscribeFuture struct {
	abstractFuture

	ReturnCodes []uint8
}

func (f *SubscribeFuture) bind(future *SubscribeFuture) {
	future.Call(func(err error) {
		if err != nil {
			f.cancel()
			return
		}

		f.ReturnCodes = future.ReturnCodes

		f.complete()
	})
}

// UnsubscribeFuture is returned by the Client on Unsubscribe.
type UnsubscribeFuture struct {
	abstractFuture
}

func (f *UnsubscribeFuture) bind(future Future) {
	future.Call(func(err error) {
		if err != nil {
			f.cancel()
			return
		}

		f.complete()
	})
}
