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
	"sync"
	"testing"
	"time"

	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

// Run runs the passed broker on a random available port and returns a channel
// that can be closed to shutdown the broker. This method is intended to be used
// in testing scenarios.
func Run(t *testing.T, engine *Engine, protocol string) (*tools.Port, chan struct{}, chan struct{}) {
	port := tools.NewPort()

	server, err := transport.Launch(port.URL(protocol))
	assert.NoError(t, err)

	quit := make(chan struct{})
	done := make(chan struct{})

	go func() {
		for {
			// get next connection and return on error
			conn, err := server.Accept()
			if err != nil {
				return
			}

			// handle connection
			engine.Handle(conn)
		}
	}()

	go func() {
		<-quit

		// check for active clients
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 0, len(engine.Clients()))

		// close broker
		engine.Close()

		// errors from close are ignored
		server.Close()

		close(done)
	}()

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
