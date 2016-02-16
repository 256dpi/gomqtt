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

import "sync"

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

/* context */

type Context struct {
	store map[string]interface{}
	mutex sync.Mutex
}

func NewContext() *Context {
	return &Context{
		store: make(map[string]interface{}),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.store[key] = value
}

func (c *Context) Get(key string) interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.store[key]
}
