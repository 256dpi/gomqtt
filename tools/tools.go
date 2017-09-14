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

// Package tools implements utilities for building MQTT 3.1.1
// (http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) components.
package tools

import "sync"

// A Counter continuously counts packet ids.
type Counter struct {
	current uint16
	mutex   sync.Mutex
}

// NewCounter returns a new counter.
func NewCounter() *Counter {
	return &Counter{
		current: 1,
	}
}

// Next will return the next id.
func (c *Counter) Next() uint16 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// cache current id
	id := c.current

	// increment id
	c.current++

	// increment again if current id is zero
	if c.current == 0 {
		c.current++
	}

	return id
}

// Reset will reset the counter.
func (c *Counter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.current = 1
}
