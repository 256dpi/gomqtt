package client

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport/flow"

	"github.com/stretchr/testify/assert"
)

func TestServicePublishSubscribe(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []uint8{0}
	suback.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := flow.New().
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

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.OfflineCallback = func() {
		close(offline)
	}

	s.MessageCallback = func(msg *packet.Message) error {
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)
		close(message)
		return nil
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online)

	assert.NoError(t, s.Subscribe("test", 0).Wait(1*time.Second))
	assert.NoError(t, s.Publish("test", []byte("test"), 0, false).Wait(1*time.Second))

	safeReceive(message)

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)
}

func TestServiceCommandsInCallback(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []uint8{0}
	suback.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	message := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)

		s.Subscribe("test", 0)
		s.Publish("test", []byte("test"), 0, false)
	}

	s.OfflineCallback = func() {
		close(offline)
	}

	s.MessageCallback = func(msg *packet.Message) error {
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)

		close(message)
		return nil
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(message)

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)
}

func TestStartStopVariations(t *testing.T) {
	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	online := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.OfflineCallback = func() {
		close(offline)
	}

	s.Start(NewConfig("tcp://localhost:" + port))
	s.Start(NewConfig("tcp://localhost:" + port)) // does nothing

	safeReceive(online)

	s.Stop(true)
	s.Stop(true) // does nothing

	safeReceive(offline)
	safeReceive(done)
}

func TestServiceUnsubscribe(t *testing.T) {
	unsubscribe := packet.NewUnsubscribePacket()
	unsubscribe.Topics = []string{"test"}
	unsubscribe.ID = 1

	unsuback := packet.NewUnsuback()
	unsuback.ID = 1

	broker := flow.New().
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

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.OfflineCallback = func() {
		close(offline)
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online)

	assert.NoError(t, s.Unsubscribe("test").Wait(1*time.Second))

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)
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

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)
		close(online)
	}

	s.OfflineCallback = func() {
		close(offline)
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online)

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)

	assert.Equal(t, 4, i)
}

func TestServiceFutureSurvival(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = "test"
	connect.CleanSession = false

	connack := connackPacket()
	connack.SessionPresent = true

	publish1 := packet.NewPublish()
	publish1.Message.Topic = "test"
	publish1.Message.Payload = []byte("test")
	publish1.Message.QOS = 1
	publish1.ID = 1

	publish2 := packet.NewPublish()
	publish2.Message.Topic = "test"
	publish2.Message.Payload = []byte("test")
	publish2.Message.QOS = 1
	publish2.Dup = true
	publish2.ID = 1

	puback := packet.NewPuback()
	puback.ID = 1

	broker1 := flow.New().
		Receive(connect).
		Send(connack).
		Receive(publish1).
		Close()

	broker2 := flow.New().
		Receive(connect).
		Send(connack).
		Receive(publish2).
		Send(puback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2)

	config := NewConfigWithClientID("tcp://localhost:"+port, "test")
	config.CleanSession = false

	s := NewService()

	s.Start(config)

	assert.NoError(t, s.Publish("test", []byte("test"), 1, false).Wait(5*time.Second))

	s.Stop(true)

	safeReceive(done)
}

func BenchmarkServicePublish(b *testing.B) {
	ready := make(chan struct{})
	done := make(chan struct{})

	c := NewService()

	c.OnlineCallback = func(_ bool) {
		close(ready)
	}

	c.OfflineCallback = func() {
		close(done)
	}

	c.Start(NewConfig("mqtt://0.0.0.0"))

	safeReceive(ready)

	for i := 0; i < b.N; i++ {
		c.Publish("test", []byte("test"), 0, false)
	}

	c.Stop(true)

	safeReceive(done)
}
