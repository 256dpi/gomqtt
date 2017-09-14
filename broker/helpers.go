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

package broker

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

// Run runs the passed broker on a random available port and returns a channel
// that can be closed to shutdown the broker. This method is intended to be used
// in testing scenarios.
func Run(t *testing.T, engine *Engine, protocol string) (string, chan struct{}, chan struct{}) {
	server, err := transport.Launch(protocol + "://localhost:0")
	assert.NoError(t, err)

	quit := make(chan struct{})
	done := make(chan struct{})

	engine.Accept(server)

	go func() {
		<-quit

		// check for active clients
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 0, len(engine.Clients()))

		// errors from close are ignored
		server.Close()

		// close broker
		engine.Close()

		// wait for proper closing
		engine.Wait(10 * time.Millisecond)

		close(done)
	}()

	_, port, _ := net.SplitHostPort(server.Addr().String())

	return port, quit, done
}

/* state */

// a state keeps track of the clients current state
type state struct {
	sync.Mutex

	current byte
}

// create new state
func newState(init byte) *state {
	return &state{
		current: init,
	}
}

// set will change to the specified state
func (s *state) set(state byte) {
	s.Lock()
	defer s.Unlock()

	s.current = state
}

// get will retrieve the current state
func (s *state) get() byte {
	s.Lock()
	defer s.Unlock()

	return s.current
}

/* Context */

// A Context is a store for custom data.
type Context struct {
	store map[string]interface{}
	mutex sync.Mutex
}

// NewContext returns a new Context.
func NewContext() *Context {
	return &Context{
		store: make(map[string]interface{}),
	}
}

// Set sets the passed value for the key in the context.
func (c *Context) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.store[key] = value
}

// Get returns the stored valued for the passed key.
func (c *Context) Get(key string) interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.store[key]
}
