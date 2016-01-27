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
	"sync"
)

// a futureStore is used to store active Futures
type futureStore struct {
	store map[uint16]Future
	mutex sync.Mutex
}

// newFutureStore will create a new futureStore
func newFutureStore() *futureStore {
	return &futureStore{
		store: make(map[uint16]Future),
	}
}

// put will store a Future
func (s *futureStore) put(id uint16, future Future) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[id] = future
}

// get will retrieve a Future
func (s *futureStore) get(id uint16) (Future) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.store[id]
}

// del will remove a Future
func (s *futureStore) del(id uint16) {
	delete(s.store, id)
}

// a packetIDGenerator generates continuously packet ids
type idGenerator struct {
	nextID uint16
	mutex sync.Mutex
}

// newIDGenerator will return a new idGenerator
func newIDGenerator() *idGenerator {
	return &idGenerator{}
}

// next will generate the next packet id
func (g *idGenerator) next() uint16 {
	g.mutex.Lock()
	defer func() {
		g.nextID++ // its safe to overflow an unsigned int
		g.mutex.Unlock()
	}()

	return g.nextID
}
