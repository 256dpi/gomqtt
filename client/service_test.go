package client

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport/flow"

	"github.com/stretchr/testify/assert"
)

func TestServicePublishSubscribe(t *testing.T) {
	subscribe := packet.NewSubscribe()
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
	subscribe := packet.NewSubscribe()
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
	unsubscribe := packet.NewUnsubscribe()
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

func TestServiceReconnectResubscribe(t *testing.T) {
	subscribe1 := packet.NewSubscribe()
	subscribe1.Subscriptions = []packet.Subscription{{Topic: "overlap/#", QOS: 0}}
	subscribe1.ID = 1

	subscribe2 := packet.NewSubscribe()
	subscribe2.Subscriptions = []packet.Subscription{{Topic: "overlap/a", QOS: 1}}
	subscribe2.ID = 2

	subscribe3 := packet.NewSubscribe()
	subscribe3.Subscriptions = []packet.Subscription{{Topic: "identical/a", QOS: 1}}
	subscribe3.ID = 3

	subscribe4 := packet.NewSubscribe()
	subscribe4.Subscriptions = []packet.Subscription{{Topic: "identical/a", QOS: 0}}
	subscribe4.ID = 4

	subscribe5 := packet.NewSubscribe()
	subscribe5.Subscriptions = []packet.Subscription{{Topic: "unsubscribe", QOS: 1}}
	subscribe5.ID = 5

	unsubscribe := packet.NewUnsubscribe()
	unsubscribe.Topics = []string{"unsubscribe"}
	unsubscribe.ID = 6

	suback1 := packet.NewSuback()
	suback1.ReturnCodes = []uint8{0}
	suback1.ID = 1

	suback2 := packet.NewSuback()
	suback2.ReturnCodes = []uint8{1}
	suback2.ID = 2

	suback3 := packet.NewSuback()
	suback3.ReturnCodes = []uint8{1}
	suback3.ID = 3

	suback4 := packet.NewSuback()
	suback4.ReturnCodes = []uint8{0}
	suback4.ID = 4

	suback5 := packet.NewSuback()
	suback5.ReturnCodes = []uint8{0}
	suback5.ID = 5

	unsuback := packet.NewUnsuback()
	unsuback.ID = 6

	resubscribe := packet.NewSubscribe()
	resubscribe.Subscriptions = []packet.Subscription{
		{Topic: "overlap/#", QOS: 0},
		{Topic: "overlap/a", QOS: 1},
		{Topic: "identical/a", QOS: 0},
	}
	resubscribe.ID = 1

	resuback := packet.NewSuback()
	resuback.ReturnCodes = []uint8{0, 1, 0}
	resuback.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	firstClose := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe1).
		Send(suback1).
		Receive(subscribe2).
		Send(suback2).
		Receive(subscribe3).
		Send(suback3).
		Receive(subscribe4).
		Send(suback4).
		Receive(subscribe5).
		Send(suback5).
		Receive(unsubscribe).
		Send(unsuback).
		Close()

	noClose := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(resubscribe).
		Send(resuback).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, firstClose, noClose)

	online1 := make(chan struct{})
	message := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()
	s.MinReconnectDelay = 50 * time.Millisecond
	s.ConnectTimeout = 50 * time.Millisecond

	i := uint32(0)

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)

		if atomic.AddUint32(&i, 1) == 1 {
			close(online1)
		}
	}

	s.OfflineCallback = func() {
		if atomic.LoadUint32(&i) == 2 {
			close(offline)
		}
	}

	s.MessageCallback = func(m *packet.Message) error {
		assert.Equal(t, "test", m.Topic)
		assert.Equal(t, []byte("test"), m.Payload)
		assert.Equal(t, uint8(0), m.QOS)
		assert.False(t, m.Retain)
		close(message)
		return nil
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online1)

	assert.NoError(t, s.SubscribeMultiple(subscribe1.Subscriptions).Wait(time.Second))
	assert.NoError(t, s.SubscribeMultiple(subscribe2.Subscriptions).Wait(time.Second))
	assert.NoError(t, s.SubscribeMultiple(subscribe3.Subscriptions).Wait(time.Second))
	assert.NoError(t, s.SubscribeMultiple(subscribe4.Subscriptions).Wait(time.Second))
	assert.NoError(t, s.SubscribeMultiple(subscribe5.Subscriptions).Wait(time.Second))
	assert.NoError(t, s.UnsubscribeMultiple(unsubscribe.Topics).Wait(time.Second))

	safeReceive(message)

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)

	assert.Equal(t, uint32(2), i)
}

func TestServiceReconnectResubscribeTimeout(t *testing.T) {
	subscribe1 := packet.NewSubscribe()
	subscribe1.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 0}}
	subscribe1.ID = 1

	suback1 := packet.NewSuback()
	suback1.ReturnCodes = []uint8{0}
	suback1.ID = 1

	broker1 := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe1).
		Send(suback1).
		Close()

	broker2 := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe1).
		End()

	broker3 := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe1).
		Send(suback1).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2, broker3)

	online1 := make(chan struct{})
	online2 := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()
	s.MinReconnectDelay = 50 * time.Millisecond
	s.ConnectTimeout = 50 * time.Millisecond
	s.ResubscribeTimeout = 55 * time.Millisecond
	s.Logger = func(msg string) {
		fmt.Println(msg)
	}

	i := uint32(0)
	s.Logger = func(msg string) {
		if msg == "Next Reconnect" {
			fmt.Println("reconnect")
			atomic.AddUint32(&i, 1)
		}
	}

	s.OnlineCallback = func(resumed bool) {
		fmt.Println("online")
		assert.False(t, resumed)
		if atomic.LoadUint32(&i) == 1 {
			close(online1)
		} else if atomic.LoadUint32(&i) == 3 {
			close(online2)
		}
	}

	s.OfflineCallback = func() {
		fmt.Println("offline")
		if atomic.LoadUint32(&i) == 3 {
			close(offline)
		}
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online1)

	assert.NoError(t, s.SubscribeMultiple(subscribe1.Subscriptions).Wait(time.Second))

	safeReceive(online2)

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)

	assert.Equal(t, uint32(3), i)
}

func TestServiceReconnectResubscribeClosed(t *testing.T) {
	subscribe1 := packet.NewSubscribe()
	subscribe1.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 0}}
	subscribe1.ID = 1

	suback1 := packet.NewSuback()
	suback1.ReturnCodes = []uint8{0}
	suback1.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker1 := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe1).
		Send(suback1).
		Close()

	broker2 := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2)

	online1 := make(chan struct{})
	message := make(chan struct{})
	offline := make(chan struct{})

	s := NewService()
	s.MinReconnectDelay = 50 * time.Millisecond
	s.ConnectTimeout = 50 * time.Millisecond
	s.ResubscribeTimeout = 55 * time.Millisecond
	s.ResubscribeAllSubscriptions = false
	s.Logger = func(msg string) {
		fmt.Println(msg)
	}
	s.MessageCallback = func(msg *packet.Message) error {
		assert.Equal(t, "test", msg.Topic)
		assert.Equal(t, []byte("test"), msg.Payload)
		assert.Equal(t, uint8(0), msg.QOS)
		assert.False(t, msg.Retain)
		close(message)
		return nil
	}

	i := uint32(0)
	s.Logger = func(msg string) {
		if msg == "Next Reconnect" {
			fmt.Println("reconnect")
			atomic.AddUint32(&i, 1)
		}
	}

	s.OnlineCallback = func(resumed bool) {
		fmt.Println("online")
		assert.False(t, resumed)
		if atomic.LoadUint32(&i) == 1 {
			close(online1)
		}
	}

	s.OfflineCallback = func() {
		fmt.Println("offline")
		if atomic.LoadUint32(&i) == 2 {
			close(offline)
		}
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online1)

	assert.NoError(t, s.SubscribeMultiple(subscribe1.Subscriptions).Wait(time.Second))

	safeReceive(message)

	s.Stop(true)

	safeReceive(offline)
	safeReceive(done)

	assert.Equal(t, uint32(2), i)
}

func TestServiceReconnectResubscribeError(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 3}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []uint8{packet.QOSFailure}
	suback.ID = 1

	broker1 := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1)

	online := make(chan struct{})

	s := NewService()
	s.MinReconnectDelay = 50 * time.Millisecond
	s.ConnectTimeout = 50 * time.Millisecond
	s.ResubscribeTimeout = 55 * time.Millisecond

	i := uint32(0)

	s.OnlineCallback = func(resumed bool) {
		assert.False(t, resumed)

		if atomic.AddUint32(&i, 1) == 1 {
			close(online)
		}
	}

	s.MessageCallback = func(msg *packet.Message) error {
		assert.FailNow(t, "MessageCallback not expected")

		return nil
	}

	s.ErrorCallback = func(err error) {
		assert.EqualError(t, err, "failed subscription")
	}

	s.Start(NewConfig("tcp://localhost:" + port))

	safeReceive(online)

	err := s.SubscribeMultiple(subscribe.Subscriptions).Wait(time.Second)
	assert.EqualError(t, err, "future canceled")

	s.Stop(true)

	safeReceive(done)

	assert.Equal(t, uint32(1), i)
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
