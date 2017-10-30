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
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/tools"
)

// A GenericFuture is returned by publish and unsubscribe methods.
type GenericFuture interface {
	// Wait will block until the future is completed or canceled. It will return
	// ErrCanceled if the future gets canceled. If the timeout is reached, an
	// ErrTimeoutExceeded is returned.
	//
	// Note: Wait will not return any Client related errors.
	Wait(timeout time.Duration) error
}

// A ConnectFuture is returned by the connect method.
type ConnectFuture interface {
	GenericFuture

	// SessionPresent will return whether a session was present.
	SessionPresent() bool

	// ReturnCode will return the connack code returned by the broker.
	ReturnCode() packet.ConnackCode
}

// A SubscribeFuture is returned by the subscribe methods.
type SubscribeFuture interface {
	GenericFuture

	// ReturnCodes will return the suback codes returned by the broker.
	ReturnCodes() []uint8
}

type genericFuture struct {
	*tools.Future
}

func newGenericFuture() *genericFuture {
	return &genericFuture{
		Future: tools.NewFuture(),
	}
}

func (f *genericFuture) Bind(f2 *genericFuture) {
	f.Future.Bind(f2.Future, nil)
}

type connectFuture struct {
	*tools.Future

	sessionPresent bool
	returnCode     packet.ConnackCode
}

func newConnectFuture() *connectFuture {
	return &connectFuture{
		Future: tools.NewFuture(),
	}
}

func (f *connectFuture) SessionPresent() bool {
	return f.sessionPresent
}

func (f *connectFuture) ReturnCode() packet.ConnackCode {
	return f.returnCode
}

type subscribeFuture struct {
	*tools.Future

	returnCodes []uint8
}

func newSubscribeFuture() *subscribeFuture {
	return &subscribeFuture{
		Future: tools.NewFuture(),
	}
}

func (f *subscribeFuture) Bind(f2 *subscribeFuture) {
	f.Future.Bind(f2.Future, func() {
		f.returnCodes = f2.returnCodes
	})
}

func (f *subscribeFuture) ReturnCodes() []uint8 {
	return f.returnCodes
}
