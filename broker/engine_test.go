package broker

import (
	"net"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"

	"github.com/stretchr/testify/assert"
)

func TestConnectTimeout(t *testing.T) {
	engine := NewEngine(NewMemoryBackend())
	engine.ConnectTimeout = 10 * time.Millisecond

	port, quit, done := Run(engine, "tcp")

	conn, err := transport.Dial("tcp://localhost:" + port)
	assert.NoError(t, err)

	pkt, err := conn.Receive()
	assert.Nil(t, pkt)
	assert.Error(t, err)

	close(quit)
	safeReceive(done)
}

func TestDefaultReadLimit(t *testing.T) {
	engine := NewEngine(NewMemoryBackend())
	engine.ReadLimit = 1

	port, quit, done := Run(engine, "tcp")

	c := client.New()
	wait := make(chan struct{})

	c.Callback = func(msg *packet.Message, err error) error {
		assert.Error(t, err)
		close(wait)
		return nil
	}

	cf, err := c.Connect(client.NewConfig("tcp://localhost:" + port))
	assert.NoError(t, err)
	assert.Error(t, cf.Wait(10*time.Second))

	safeReceive(wait)
	close(quit)
	safeReceive(done)
}

type pendingServer struct {
	waiting chan struct{}
	conns   chan transport.Conn
}

func (s *pendingServer) Accept() (transport.Conn, error) {
	s.waiting <- struct{}{}
	conn, ok := <-s.conns
	if !ok {
		return nil, net.ErrClosed
	}
	return conn, nil
}

func (s *pendingServer) Close() error {
	return nil
}

func (s *pendingServer) Addr() net.Addr {
	return nil
}

func TestEngineCloseWhileHandling(t *testing.T) {
	engine := NewEngine(NewMemoryBackend())

	server := &pendingServer{
		waiting: make(chan struct{}),
		conns:   make(chan transport.Conn),
	}
	engine.Accept(server)
	<-server.waiting

	// close the engine while a connection is accepted
	done := make(chan struct{})
	go func() {
		engine.Close()
		close(done)
	}()
	<-engine.tomb.Dying()

	local, remote := net.Pipe()
	defer remote.Close()
	server.conns <- transport.NewNetConn(local)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("close did not return")
	}

	// the accepted connection is closed
	_, err := remote.Read(make([]byte, 1))
	assert.Error(t, err)
}
