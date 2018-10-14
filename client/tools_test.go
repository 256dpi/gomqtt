package client

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport/flow"

	"github.com/stretchr/testify/assert"
)

func TestClearSession(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = "test"

	broker := flow.New().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := ClearSession(NewConfigWithClientID("tcp://localhost:"+port, "test"), 1*time.Second)
	assert.NoError(t, err)

	safeReceive(done)
}

func TestClearRetainedMessage(t *testing.T) {
	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = nil
	publish.Message.Retain = true

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := ClearRetainedMessage(NewConfig("tcp://localhost:"+port), "test", 1*time.Second)
	assert.NoError(t, err)

	safeReceive(done)
}

func TestPublishMessage(t *testing.T) {
	publish := packet.NewPublish()
	publish.Message = packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		Retain:  true,
	}

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := PublishMessage(NewConfig("tcp://localhost:"+port), &publish.Message, 1*time.Second)
	assert.NoError(t, err)

	safeReceive(done)
}

func TestReceiveMessage(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.ID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test"},
	}

	suback := packet.NewSuback()
	suback.ID = 1
	suback.ReturnCodes = []packet.QOS{0}

	publish := packet.NewPublish()
	publish.Message = packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		Retain:  true,
	}

	broker := flow.New().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	msg, err := ReceiveMessage(NewConfig("tcp://localhost:"+port), "test", 0, 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, publish.Message.String(), msg.String())

	safeReceive(done)
}
