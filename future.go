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

// Future represents information that might become available in the future.
type Future interface {
	// Wait will block until the future is completed. If a timeout is specified
	// it might return a ErrTimeoutExceeded.
	Wait(timeout ...time.Duration) error

	// Completed returns true if the future is completed.
	Completed() bool
}

type abstractFuture struct {
	channel chan struct{}
}

func (f *abstractFuture) initialize() {
	f.channel = make(chan struct{})
}

func (f *abstractFuture) Wait(timeout ...time.Duration) error {
	if len(timeout) > 0 {
		select {
		case <-f.channel:
			return nil
		case <-time.After(timeout[0]):
			return ErrTimeoutExceeded
		}
	}

	<-f.channel
	return nil
}

func (f *abstractFuture) complete() {
	close(f.channel)
}

func (f *abstractFuture) Completed() bool {
	select {
	case <-f.channel:
		return true
	default:
		return false
	}
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
