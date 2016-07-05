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

	"github.com/fortytw2/leaktest"
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/stretchr/testify/assert"
)

func TestClearSession(t *testing.T) {
	defer leaktest.Check(t)()

	connect := connectPacket()
	connect.ClientID = "test"

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	ClearSession(port.URL(), "test")

	<-done
}

func TestClearRetainedMessage(t *testing.T) {
	defer leaktest.Check(t)()

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = nil
	publish.Message.Retain = true

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	ClearRetainedMessage(port.URL(), "test")

	<-done
}

func TestServicePublishSubscribe(t *testing.T) {
	defer leaktest.Check(t)()

	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test"},
	}
	subscribe.PacketID = 1

	suback := packet.NewSubackPacket()
	suback.ReturnCodes = []uint8{0}
	suback.PacketID = 1

	publish := packet.NewPublishPacket()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

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

	s.Message = func(msg *packet.Message) {
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)
		close(message)
	}

	s.Start(port.URL(), nil)

	<-online

	s.Subscribe("test", 0, false).Wait()
	s.Publish("test", []byte("test"), 0, false)

	<-message

	s.Stop(true)

	<-offline
	<-done
}

func TestStartStopVariations(t *testing.T) {
	defer leaktest.Check(t)()

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

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

	s.Start(port.URL(), nil)
	s.Start(port.URL(), nil) // <- does nothing

	<-online

	s.Stop(true)
	s.Stop(true) // <- does nothing

	<-offline
	<-done
}

func TestServiceUnsubscribe(t *testing.T) {
	defer leaktest.Check(t)()

	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = []string{"test"}
	unsubscribe.PacketID = 1

	unsuback := packet.NewUnsubackPacket()
	unsuback.PacketID = 1

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(unsubscribe).
		Send(unsuback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

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

	s.Start(port.URL(), nil)

	<-online

	s.Unsubscribe("test", false).Wait()

	s.Stop(true)

	<-offline
	<-done
}

func TestServiceReconnect(t *testing.T) {
	defer leaktest.Check(t)()

	delay := tools.NewFlow().
		Receive(connectPacket()).
		Delay(55 * time.Millisecond).
		End()

	noDelay := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, delay, delay, delay, noDelay)

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

	s.Start(port.URL(), nil)

	<-online

	s.Stop(true)

	<-offline
	<-done

	assert.Equal(t, 4, i)
}

func TestServiceFutureSurvival(t *testing.T) {
	defer leaktest.Check(t)()

	connect := connectPacket()
	connect.ClientID = "test"
	connect.CleanSession = false

	connack := connackPacket()
	connack.SessionPresent = true

	publish1 := packet.NewPublishPacket()
	publish1.Message.Topic = "test"
	publish1.Message.Payload = []byte("test")
	publish1.Message.QOS = 1
	publish1.PacketID = 1

	publish2 := packet.NewPublishPacket()
	publish2.Message.Topic = "test"
	publish2.Message.Payload = []byte("test")
	publish2.Message.QOS = 1
	publish2.Dup = true
	publish2.PacketID = 1

	puback := packet.NewPubackPacket()
	puback.PacketID = 1

	broker1 := tools.NewFlow().
		Receive(connect).
		Send(connack).
		Receive(publish1).
		Close()

	broker2 := tools.NewFlow().
		Receive(connect).
		Send(connack).
		Receive(publish2).
		Send(puback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2)

	options := NewOptions()
	options.ClientID = "test"
	options.CleanSession = false

	s := NewService()

	s.Start(port.URL(), options)

	err := s.Publish("test", []byte("test"), 1, false).Wait()
	assert.NoError(t, err)

	s.Stop(true)

	<-done
}
