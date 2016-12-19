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

package transport

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
)

type webSocketStream struct {
	conn   *websocket.Conn
	reader io.Reader
}

// ErrNotBinary may be returned by WebSocket connection when a message is
// received that is not binary.
var ErrNotBinary = errors.New("received web socket message is not binary")

func (s *webSocketStream) Read(p []byte) (int, error) {
	total := 0
	buf := p

	for {
		// get next reader
		if s.reader == nil {
			messageType, reader, err := s.conn.NextReader()
			if _, ok := err.(*websocket.CloseError); ok {
				return 0, io.EOF
			} else if err != nil {
				return 0, err
			} else if messageType != websocket.BinaryMessage {
				return 0, ErrNotBinary
			}

			// set current reader
			s.reader = reader
		}

		// read data
		n, err := s.reader.Read(buf)

		// increment counter
		total += n
		buf = buf[n:]

		// handle EOF
		if err == io.EOF {
			// clear reader
			s.reader = nil

			continue
		}

		// handle other errors
		if err != nil {
			return total, err
		}

		return total, err
	}
}

func (s *webSocketStream) Write(p []byte) (n int, err error) {
	// create writer if missing
	writer, err := s.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	// write packet to writer
	n, err = writer.Write(p)
	if err != nil {
		return n, err
	}

	// close temporary writer
	err = writer.Close()
	if err != nil {
		return n, err
	}

	return n, nil
}

var closeMessage = websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")

// The WebSocketConn wraps a gorilla WebSocket.Conn. The implementation supports
// packets that are chunked over several WebSocket messages and packets that are
// coalesced to one WebSocket message.
type WebSocketConn struct {
	conn   *websocket.Conn
	stream *packet.Stream

	flushTimer *time.Timer
	flushError error

	sMutex sync.Mutex
	rMutex sync.Mutex

	readTimeout time.Duration
}

// NewWebSocketConn returns a new WebSocketConn.
func NewWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	s := &webSocketStream{
		conn: conn,
	}

	return &WebSocketConn{
		conn: conn,
		stream: packet.NewStream(s, s),
	}
}

// LocalAddr returns the local network address.
func (c *WebSocketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *WebSocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Send will write the packet to the underlying connection. It will return
// an Error if there was an error while encoding or writing to the
// underlying connection.
//
// Note: Only one goroutine can Send at the same time.
func (c *WebSocketConn) Send(pkt packet.Packet) error {
	c.sMutex.Lock()
	defer c.sMutex.Unlock()

	// write packet
	err := c.write(pkt)
	if err != nil {
		return err
	}

	// stop the timer if existing
	if c.flushTimer != nil {
		c.flushTimer.Stop()
	}

	// flush buffer
	return c.flush()
}

// BufferedSend will write the packet to an internal buffer. It will flush
// the internal buffer automatically when it gets stale. Encoding errors are
// directly returned as in Send, but any network errors caught while flushing
// the buffer at a later time will be returned on the next call.
//
// Note: Only one goroutine can call BufferedSend at the same time.
func (c *WebSocketConn) BufferedSend(pkt packet.Packet) error {
	c.sMutex.Lock()
	defer c.sMutex.Unlock()

	// create the timer if missing
	if c.flushTimer == nil {
		c.flushTimer = time.AfterFunc(flushTimeout, c.asyncFlush)
		c.flushTimer.Stop()
	}

	// return any error from asyncFlush
	if c.flushError != nil {
		return c.flushError
	}

	// write packet
	err := c.write(pkt)
	if err != nil {
		return err
	}

	// queue asyncFlush
	c.flushTimer.Reset(flushTimeout)

	return nil
}

func (c *WebSocketConn) write(pkt packet.Packet) error {
	err := c.stream.Write(pkt)
	if err != nil {
		c.end()

		return err
	}

	return nil
}

func (c *WebSocketConn) flush() error {
	err := c.stream.Flush()
	if err != nil {
		c.end()

		return err
	}

	return nil
}

func (c *WebSocketConn) asyncFlush() {
	c.sMutex.Lock()
	defer c.sMutex.Unlock()

	// flush buffer and save an eventual error
	err := c.flush()
	if err != nil {
		c.flushError = err
	}
}

// Receive will read from the underlying connection and return a fully read
// packet. It will return an Error if there was an error while decoding or
// reading from the underlying connection.
//
// Note: Only one goroutine can Receive at the same time.
func (c *WebSocketConn) Receive() (packet.Packet, error) {
	c.rMutex.Lock()
	defer c.rMutex.Unlock()

	// read next packet
	pkt, err := c.stream.Read()
	if err != nil {
		c.end()
		return nil, err
	}

	// reset timeout
	c.resetTimeout()

	return pkt, nil
}

func (c *WebSocketConn) end() error {
	// write close message
	err := c.conn.WriteMessage(websocket.CloseMessage, closeMessage)
	if err != nil {
		return err
	}

	return c.conn.Close()
}

// Close will close the underlying connection and cleanup resources. It will
// return an Error if there was an error while closing the underlying
// connection.
func (c *WebSocketConn) Close() error {
	c.sMutex.Lock()
	defer c.sMutex.Unlock()

	return c.end()
}

// SetReadLimit sets the maximum size of a packet that can be received.
// If the limit is greater than zero, Receive will close the connection and
// return an Error if receiving the next packet will exceed the limit.
func (c *WebSocketConn) SetReadLimit(limit int64) {
	c.stream.Decoder.Limit = limit
}

// SetReadTimeout sets the maximum time that can pass between reads.
// If no data is received in the set duration the connection will be closed
// and Read returns an error.
func (c *WebSocketConn) SetReadTimeout(timeout time.Duration) {
	c.readTimeout = timeout
	c.resetTimeout()
}

// UnderlyingConn returns the underlying websocket.Conn.
func (c *WebSocketConn) UnderlyingConn() *websocket.Conn {
	return c.conn
}

func (c *WebSocketConn) resetTimeout() {
	if c.readTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	} else {
		c.conn.SetReadDeadline(time.Time{})
	}
}
