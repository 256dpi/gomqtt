package client

import (
	"net"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/tools"
	"github.com/256dpi/gomqtt/transport"
	"github.com/stretchr/testify/assert"
)

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(1 * time.Minute):
		panic("nothing received")
	case <-ch:
	}
}

func errorCallback(t *testing.T) func(*packet.Message, error) error {
	return func(msg *packet.Message, err error) error {
		if err != nil {
			println(err.Error())
		}

		assert.Fail(t, "callback should not have been called")
		return nil
	}
}

func fakeBroker(t *testing.T, testFlows ...*tools.Flow) (chan struct{}, string) {
	done := make(chan struct{})

	server, err := transport.Launch("tcp://localhost:0")
	assert.NoError(t, err)

	go func() {
		for _, flow := range testFlows {
			conn, err := server.Accept()
			assert.NoError(t, err)

			err = flow.Test(conn)
			assert.NoError(t, err)
		}

		err = server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	_, port, _ := net.SplitHostPort(server.Addr().String())

	return done, port
}

func connectPacket() *packet.ConnectPacket {
	pkt := packet.NewConnectPacket()
	pkt.CleanSession = true
	pkt.KeepAlive = 30
	return pkt
}

func connackPacket() *packet.ConnackPacket {
	pkt := packet.NewConnackPacket()
	pkt.ReturnCode = packet.ConnectionAccepted
	pkt.SessionPresent = false
	return pkt
}

func disconnectPacket() *packet.DisconnectPacket {
	return packet.NewDisconnectPacket()
}
