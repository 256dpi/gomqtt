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
	"testing"
	"errors"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/gomqtt/packet"
)

func testOptions() *Options {
	return NewOptions("test/" + uuid.NewV4().String())
}

func errorCallback(t *testing.T) func(*Message, error) {
	return func(msg *Message, err error) {
		assert.Fail(t, "callback should not have been called")
	}
}

type errorStore struct {
	MemoryStore

	put   bool
	get   bool
	del   bool
	all   bool
	reset bool
}

func (s *errorStore) Put(dir string, pkt packet.Packet) error {
	if s.put {
		return errors.New("error")
	}

	return s.MemoryStore.Put(dir, pkt)
}

func (s *errorStore) Get(dir string, id uint16) (packet.Packet, error) {
	if s.get {
		return nil, errors.New("error")
	}

	return s.MemoryStore.Get(dir, id)
}

func (s *errorStore) Del(dir string, id uint16) error {
	if s.del {
		return errors.New("error")
	}

	return s.MemoryStore.Del(dir, id)
}

func (s *errorStore) All(dir string) ([]packet.Packet, error) {
	if s.all {
		return nil, errors.New("error")
	}

	return s.MemoryStore.All(dir)
}

func (s *errorStore) Reset() error {
	if s.reset {
		return errors.New("error")
	}

	return s.MemoryStore.Reset()
}
