package flow

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"

	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {
	connect := packet.NewConnect()
	connack := packet.NewConnack()

	subscribe := packet.NewSubscribePacket()
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: "test"},
	}
	subscribe.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"

	wait := make(chan struct{})
	cb := func() {
		close(wait)
	}

	server := New().
		Receive(connect).
		Send(connack).
		Run(cb).
		Skip(&packet.SubscribePacket{}).
		Receive(publish).
		Close()

	client := New().
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
