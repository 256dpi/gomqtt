package client

import (
	"testing"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/stretchr/testify/assert"
)

func TestClearSession(t *testing.T) {
	connect := connectPacket()
	connect.ClientID = "test"

	broker := tools.NewFlow().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := ClearSession(NewConfigWithClientID("tcp://localhost:"+port, "test"))
	assert.NoError(t, err)

	<-done
}

func TestClearRetainedMessage(t *testing.T) {
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

	err := ClearRetainedMessage(NewConfig("tcp://localhost:"+port), "test")
	assert.NoError(t, err)

	<-done
}

func TestPublishMessage(t *testing.T) {
	publish := packet.NewPublishPacket()
	publish.Message = packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		Retain:  true,
	}

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	err := PublishMessage(NewConfig("tcp://localhost:"+port), &publish.Message)
	assert.NoError(t, err)

	<-done
}

func TestReceiveMessage(t *testing.T) {
	subscribe := packet.NewSubscribePacket()
	subscribe.PacketID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test"},
	}

	suback := packet.NewSubackPacket()
	suback.PacketID = 1
	suback.ReturnCodes = []uint8{0}

	publish := packet.NewPublishPacket()
	publish.Message = packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
		Retain:  true,
	}

	broker := tools.NewFlow().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	msg, err := ReceiveMessage(NewConfig("tcp://localhost:"+port), "test", 0)
	assert.NoError(t, err)
	assert.Equal(t, publish.Message.String(), msg.String())

	<-done
}
