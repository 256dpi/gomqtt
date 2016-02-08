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

	"github.com/gomqtt/flow"
	"github.com/gomqtt/packet"
)

func TestClearSession(t *testing.T){
	connect := connectPacket()
	connect.ClientID = []byte("test")

	broker := flow.New().
		Receive(connect).
		Send(connackPacket(packet.ConnectionAccepted)).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	ClearSession(tp.url(), "test")

	<-done
}

// -- client
// should attempt to reconnect once server is down
// should reconnect to multiple host-ports combination if servers is passed
// should reconnect if a connack is not received in an interval
// shoud not be cleared by the connack timer
// shoud not keep requeueing the first message when offline
// should queue message until connected
// should publish a message (offline)
// should send a subscribe message (offline)
// should send an unsubscribe packet (offline)
// should reconnect after stream disconnect
// should emit \'reconnect\' when reconnecting
// should emit \'offline\' after going offline
// should not reconnect if it was ended by the user
// should setup a reconnect timer on disconnect
// should allow specification of a reconnect period
