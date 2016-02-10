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
	"time"

	"github.com/gomqtt/flow"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestClearSession(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = []byte("test")

	broker := flow.New().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	ClearSession(tp.url(), "test")

	<-done
}

func TestServicePublishSubscribe(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: []byte("test")},
	}
	subscribe.PacketID = 0

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []byte{0}
	suback.PacketID = 0

	publish := packet.NewPublishPacket()
	publish.Topic = []byte("test")
	publish.Payload = []byte("test")

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	online := make(chan struct{})
	message := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()

	s.Online = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.Offline = func() {
		close(offline)
	}

	s.Message = func(topic string, payload []byte) {
		assert.Equal(t, "test", topic)
		assert.Equal(t, []byte("test"), payload)
		close(message)
	}

	s.Start(tp.url(), nil)

	<-online

	s.Subscribe("test", 0).Wait()
	s.Publish("test", []byte("test"), 0, false)

	<-message

	s.Stop(true)

	<-offline
	<-done
}

func TestStartStopVariations(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	online := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()

	s.Online = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.Offline = func() {
		close(offline)
	}

	s.Start(tp.url(), nil)
	s.Start(tp.url(), nil) // <- does nothing

	<-online

	s.Stop(true)
	s.Stop(true) // <- does nothing

	<-offline
	<-done
}

func TestServiceUnsubscribe(t *testing.T) {
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = [][]byte{[]byte("test")}
	unsubscribe.PacketID = 0

	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = 0

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(unsubscribe).
		Send(unsuback).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, broker)

	online := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()

	s.Online = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.Offline = func() {
		close(offline)
	}

	s.Start(tp.url(), nil)

	<-online

	s.Unsubscribe("test").Wait()

	s.Stop(true)

	<-offline
	<-done
}

func TestServiceReconnect(t *testing.T) {
	delay := flow.New().
		Receive(connectPacket()).
		Delay(55 * time.Millisecond).
		End()

	noDelay := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, tp := fakeBroker(t, delay, delay, delay, noDelay)

	online := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()
	s.MinReconnectDelay = 50 * time.Millisecond
	s.ConnectTimeout = 50 * time.Millisecond

	i := 0
	s.Logger = func(msg string) {
		if msg == "Next Reconnect" {
			i++
		}
	}

	s.Online = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.Offline = func() {
		close(offline)
	}

	s.Start(tp.url(), nil)

	<-online

	s.Stop(true)

	<-offline
	<-done

	assert.Equal(t, 4, i)
}
