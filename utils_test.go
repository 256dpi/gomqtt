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
	"github.com/gomqtt/session"
)

func testOptions() *Options {
	return NewOptions("test/" + uuid.NewV4().String())
}

func errorCallback(t *testing.T) func(*Message, error) {
	return func(msg *Message, err error) {
		assert.Fail(t, "callback should not have been called")
	}
}

type testSession struct {
	session.MemorySession

	saveError   bool
	lookupError bool
	deleteError bool
	allError    bool
	resetError  bool
}

func (s *testSession) SavePacket(direction string, pkt packet.Packet) error {
	if s.saveError {
		return errors.New("error")
	}

	return s.MemorySession.SavePacket(direction, pkt)
}

func (s *testSession) LookupPacket(direction string, id uint16) (packet.Packet, error) {
	if s.lookupError {
		return nil, errors.New("error")
	}

	return s.MemorySession.LookupPacket(direction, id)
}

func (s *testSession) DeletePacket(direction string, id uint16) error {
	if s.deleteError {
		return errors.New("error")
	}

	return s.MemorySession.DeletePacket(direction, id)
}

func (s *testSession) AllPackets(direction string) ([]packet.Packet, error) {
	if s.allError {
		return nil, errors.New("error")
	}

	return s.MemorySession.AllPackets(direction)
}

func (s *testSession) Reset() error {
	if s.resetError {
		return errors.New("error")
	}

	return s.MemorySession.Reset()
}
