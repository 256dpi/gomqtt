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

	err := ClearSession(NewOptionsWithClientID(port.URL(), "test"))
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

	err := ClearRetainedMessage(NewOptions(port.URL()), "test")
	assert.NoError(t, err)

	<-done
}
