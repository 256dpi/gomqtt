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
	"testing"
	"time"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {
	connect := packet.NewConnectPacket()
	connack := packet.NewConnackPacket()

	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test"},
	}
	subscribe.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"

	wait := make(chan struct{})
	cb := func() {
		close(wait)
	}

	server := NewFlow().
		Receive(connect).
		Send(connack).
		Run(cb).
		Skip().
		Receive(publish).
		Close()

	client := NewFlow().
		Send(connect).
		Receive(connack).
		Wait(wait).
		Send(subscribe).
		Send(publish).
		Delay(5 * time.Millisecond).
		End()

	pipe := NewPipe()

	errCh := server.TestAsync(pipe, 100*time.Millisecond)

	err := client.Test(pipe)
	assert.NoError(t, err)

	err = <-errCh
	assert.NoError(t, err)
}

func TestAlreadyClosedError(t *testing.T) {
	pipe := NewPipe()
	pipe.Close()

	err := pipe.Send(nil)
	assert.Error(t, err)
}
