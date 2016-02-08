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
	"fmt"
	"net"
	"testing"

	"github.com/gomqtt/flow"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/session"
	"github.com/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func errorCallback(t *testing.T) func(string, []byte, error) {
	return func(topic string, payload []byte, err error) {
		if err != nil {
			println(err.Error())
		}

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

// the testPort
type testPort int

// returns a new testPort
func newTestPort() *testPort {
	// taken from: https://github.com/phayes/freeport/blob/master/freeport.go

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	p := testPort(l.Addr().(*net.TCPAddr).Port)
	return &p
}

// generates the url for that testPort
func (p *testPort) url() string {
	return fmt.Sprintf("tcp://localhost:%d/", int(*p))
}

// generates a protected url for that testPort
func (p *testPort) protectedURL(user, password string) string {
	return fmt.Sprintf("tcp://%s:%s@localhost:%d/", user, password, int(*p))
}

func fakeBroker(t *testing.T, testFlow *flow.Flow) (chan struct{}, *testPort) {
	tp := newTestPort()
	done := make(chan struct{})

	server, err := transport.Launch(tp.url())
	assert.NoError(t, err)

	go func() {
		conn, err := server.Accept()
		assert.NoError(t, err)

		testFlow.Test(t, conn)

		err = server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	return done, tp
}

func connectPacket() *packet.ConnectPacket {
	pkt := packet.NewConnectPacket()
	pkt.CleanSession = true
	pkt.KeepAlive = 30
	return pkt
}

func connackPacket(returnCode packet.ConnackCode) *packet.ConnackPacket {
	pkt := packet.NewConnackPacket()
	pkt.ReturnCode = returnCode
	pkt.SessionPresent = false
	return pkt
}

func disconnectPacket() *packet.DisconnectPacket {
	return packet.NewDisconnectPacket()
}
