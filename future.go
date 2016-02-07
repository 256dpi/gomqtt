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

// ErrTimeoutExceeded is returned by Wait if the specified timeout is exceeded.
var ErrTimeoutExceeded = errors.New("timeout exceeded")

// ErrCanceled is returned by Wait if the future gets canceled while waiting.
var ErrCanceled = errors.New("canceled")

// Future represents information that might become available in the future.
type Future interface {
	// Wait will block until the future is completed or canceled. It will return
	// ErrCanceled  if the future gets canceled. If a timeout is specified it
	// might return a ErrTimeoutExceeded.
	Wait(timeout ...time.Duration) error

	// Call calls the supplied callback in a separate goroutine when Wait returns.
	Call(func(err error), ...time.Duration)

	complete()
	cancel()
}

type abstractFuture struct {
	completeChannel chan struct{}
	cancelChannel  chan struct{}
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
			return ErrCanceled
		case <-time.After(timeout[0]):
			return ErrTimeoutExceeded
		}
	}

	select {
	case <-f.completeChannel:
		return nil
	case <-f.cancelChannel:
		return ErrCanceled
	}
}

func (f *abstractFuture) Call(callback func(error), timeout ...time.Duration) {
	go func(){
		callback(f.Wait(timeout...))
	}()
}

func (f *abstractFuture) complete() {
	close(f.completeChannel)
}

func (f *abstractFuture) cancel() {
	close(f.cancelChannel)
}

// ConnectFuture is returned by the Client on Connect.
type ConnectFuture struct {
	abstractFuture

	SessionPresent bool
	ReturnCode     packet.ConnackCode
}

// PublishFuture is returned by the Client on Publish.
type PublishFuture struct {
	abstractFuture
}

// SubscribeFuture is returned by the Client on Subscribe.
type SubscribeFuture struct {
	abstractFuture

	ReturnCodes []byte
}

// UnsubscribeFuture is returned by the Client on Unsubscribe.
type UnsubscribeFuture struct {
	abstractFuture
}
