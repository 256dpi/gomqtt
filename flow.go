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

package tools

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

// A Conn defines an abstract interface for connections used with a Flow.
type Conn interface {
	Send(pkt packet.Packet) error
	Receive() (packet.Packet, error)
	Close() error
}

// The Pipe pipes packets from Send to Receive.
type Pipe struct {
	pipe  chan packet.Packet
	close chan struct{}
}

// NewPipe returns a new Pipe.
func NewPipe() *Pipe {
	return &Pipe{
		pipe:  make(chan packet.Packet),
		close: make(chan struct{}),
	}
}

// Send returns packet on next Receive call.
func (conn *Pipe) Send(pkt packet.Packet) error {
	select {
	case conn.pipe <- pkt:
		return nil
	case <-conn.close:
		return errors.New("already closed")
	}
}

// Receive returns the packet being sent with Send.
func (conn *Pipe) Receive() (packet.Packet, error) {
	select {
	case pkt := <-conn.pipe:
		return pkt, nil
	case <-conn.close:
		return nil, io.EOF
	}
}

// Close will close the conn and let Send and Receive return errors.
func (conn *Pipe) Close() error {
	close(conn.close)
	return nil
}

// All available action types.
const (
	actionSend byte = iota
	actionReceive
	actionSkip
	actionWait
	actionRun
	actionDelay
	actionClose
	actionEnd
)

// An Action is a step in a flow.
type action struct {
	kind     byte
	packet   packet.Packet
	fn       func()
	ch       chan struct{}
	duration time.Duration
}

// A Flow is a sequence of actions that can be tested against a connection.
type Flow struct {
	actions []*action
}

// NewFlow returns a new flow.
func NewFlow() *Flow {
	return &Flow{
		actions: make([]*action, 0),
	}
}

// Send will send and one packet.
func (f *Flow) Send(pkt packet.Packet) *Flow {
	f.add(&action{
		kind:   actionSend,
		packet: pkt,
	})

	return f
}

// Receive will receive and match one packet.
func (f *Flow) Receive(pkt packet.Packet) *Flow {
	f.add(&action{
		kind:   actionReceive,
		packet: pkt,
	})

	return f
}

// Skip will receive one packet without matching it.
func (f *Flow) Skip() *Flow {
	f.add(&action{
		kind: actionSkip,
	})

	return f
}

// Wait will wait until the specified channel is closed.
func (f *Flow) Wait(ch chan struct{}) *Flow {
	f.add(&action{
		kind: actionWait,
		ch:   ch,
	})

	return f
}

// Run will call the supplied function and wait until it returns.
func (f *Flow) Run(fn func()) *Flow {
	f.add(&action{
		kind: actionRun,
		fn:   fn,
	})

	return f
}

// Delay will suspend the flow using the specified duration.
func (f *Flow) Delay(d time.Duration) *Flow {
	f.add(&action{
		kind:     actionDelay,
		duration: d,
	})

	return f
}

// Close will immediately close the connection.
func (f *Flow) Close() *Flow {
	f.add(&action{
		kind: actionClose,
	})

	return f
}

// End will match proper connection close.
func (f *Flow) End() *Flow {
	f.add(&action{
		kind: actionEnd,
	})

	return f
}

// Test starts the flow on the given Conn and reports to the specified test.
func (f *Flow) Test(t *testing.T, conn Conn) {
	for _, action := range f.actions {
		switch action.kind {
		case actionSend:
			err := conn.Send(action.packet)
			assert.NoError(t, err)
		case actionReceive:
			pkt, err := conn.Receive()
			assert.NoError(t, err)
			assert.NotNil(t, pkt)

			if pkt != nil {
				assert.Equal(t, action.packet.String(), pkt.String())
			}
		case actionSkip:
			_, err := conn.Receive()
			assert.NoError(t, err)
		case actionWait:
			<-action.ch
		case actionRun:
			action.fn()
		case actionDelay:
			time.Sleep(action.duration)
		case actionClose:
			err := conn.Close()
			assert.NoError(t, err)
		case actionEnd:
			_, err := conn.Receive()
			if err != nil {
				assert.Contains(t, err.Error(), "EOF")
			}
		}
	}
}

// add will add the specified action.
func (f *Flow) add(action *action) {
	f.actions = append(f.actions, action)
}
