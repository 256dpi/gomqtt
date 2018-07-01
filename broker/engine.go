// Package broker implements an extensible MQTT broker.
package broker

import (
	"net"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"gopkg.in/tomb.v2"
)

// LogEvent are received by a Logger.
type LogEvent int

const (
	// NewConnection is emitted when a client comes online.
	NewConnection LogEvent = iota

	// PacketReceived is emitted when a packet has been received.
	PacketReceived

	// MessagePublished is emitted after a message has been published.
	MessagePublished

	// MessageForwarded is emitted after a message has been forwarded.
	MessageForwarded

	// PacketSent is emitted when a packet has been sent.
	PacketSent

	// LostConnection is emitted when the connection has been terminated.
	LostConnection

	// TransportError is emitted when an underlying transport error occurs.
	TransportError

	// SessionError is emitted when a call to the session fails.
	SessionError

	// BackendError is emitted when a call to the backend fails.
	BackendError

	// ClientError is emitted when the client violates the protocol.
	ClientError
)

// The Logger callback handles incoming log messages.
type Logger func(LogEvent, *Client, packet.GenericPacket, *packet.Message, error)

// The Engine handles incoming connections and connects them to the backend.
type Engine struct {
	Backend Backend
	Logger  Logger

	ConnectTimeout   time.Duration
	DefaultReadLimit int64
	DefaultReadBuffer int
	DefaultWriteBuffer int

	closing   bool
	clients   []*Client
	mutex     sync.Mutex
	waitGroup sync.WaitGroup

	tomb tomb.Tomb
}

// NewEngine returns a new Engine with a basic MemoryBackend.
func NewEngine() *Engine {
	return NewEngineWithBackend(NewMemoryBackend())
}

// NewEngineWithBackend returns a new Engine with a custom Backend.
func NewEngineWithBackend(backend Backend) *Engine {
	return &Engine{
		Backend:        backend,
		ConnectTimeout: 10 * time.Second,
		clients:        make([]*Client, 0),
	}
}

// Accept begins accepting connections from the passed server.
func (e *Engine) Accept(server transport.Server) {
	e.tomb.Go(func() error {
		for {
			conn, err := server.Accept()
			if err != nil {
				return err
			}

			if !e.Handle(conn) {
				return nil
			}
		}
	})
}

// Handle takes over responsibility and handles a transport.Conn. It returns
// false if the engine is closing and the connection has been closed.
func (e *Engine) Handle(conn transport.Conn) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// check conn
	if conn == nil {
		panic("passed conn is nil")
	}

	// close conn immediately when closing
	if e.closing {
		conn.Close()
		return false
	}

	// set default read limit
	conn.SetReadLimit(e.DefaultReadLimit)

	// TODO: Buffers should be configured before the socket is opened.
	// go1.11 should provide a custom dialer facility that might allow this

	// set default buffer sizes
	conn.SetBuffers(e.DefaultReadBuffer, e.DefaultWriteBuffer)

	// set initial read timeout
	conn.SetReadTimeout(e.ConnectTimeout)

	// handle client
	newClient(e, conn)

	return true
}

// Clients returns a current list of connected clients.
func (e *Engine) Clients() []*Client {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// copy list
	clients := make([]*Client, len(e.clients))
	copy(clients, e.clients)

	return clients
}

// Close will stop handling incoming connections and close all current clients.
// The call will block until all clients are properly closed.
//
// Note: All passed servers to Accept must be closed before calling this method.
func (e *Engine) Close() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// set closing
	e.closing = true

	// stop acceptors
	e.tomb.Kill(nil)
	e.tomb.Wait()

	// close all clients
	for _, client := range e.clients {
		client.Close(false)
	}
}

// Wait can be called after close to wait until all clients have been closed.
// The method returns whether all clients have been closed (true) or the timeout
// has been reached (false).
func (e *Engine) Wait(timeout time.Duration) bool {
	wait := make(chan struct{})

	go func() {
		e.waitGroup.Wait()
		close(wait)
	}()

	select {
	case <-wait:
		return true
	case <-time.After(timeout):
		return false
	}
}

// clients call add to add themselves to the list
func (e *Engine) add(client *Client) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// add client
	e.clients = append(e.clients, client)

	// increment wait group
	e.waitGroup.Add(1)
}

// clients call remove when closed to remove themselves from the list
func (e *Engine) remove(client *Client) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// get index of client
	index := 0
	for i, c := range e.clients {
		if c == client {
			index = i
			break
		}
	}

	// remove client from list
	// https://github.com/golang/go/wiki/SliceTricks
	e.clients[index] = e.clients[len(e.clients)-1]
	e.clients[len(e.clients)-1] = nil
	e.clients = e.clients[:len(e.clients)-1]

	// decrement wait group
	e.waitGroup.Add(-1)
}

// Run runs the passed engine on a random available port and returns a channel
// that can be closed to shutdown the engine. This method is intended to be used
// in testing scenarios.
func Run(engine *Engine, protocol string) (string, chan struct{}, chan struct{}) {
	// launch server
	server, err := transport.Launch(protocol + "://localhost:0")
	if err != nil {
		panic(err)
	}

	// prepare channels
	quit := make(chan struct{})
	done := make(chan struct{})

	// start accepting connections
	engine.Accept(server)

	// prepare shutdown
	go func() {
		// wait for signal
		<-quit

		// errors from close are ignored
		server.Close()

		// close broker
		engine.Close()

		// wait for proper closing
		engine.Wait(10 * time.Millisecond)

		close(done)
	}()

	// get random port
	_, port, _ := net.SplitHostPort(server.Addr().String())

	return port, quit, done
}
