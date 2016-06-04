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
	"bytes"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
)

// The WebSocketConn wraps a gorilla WebSocket.Conn.
type WebSocketConn struct {
	conn *websocket.Conn

	flushTimer *time.Timer
	flushError error

	sMutex sync.Mutex
	rMutex sync.Mutex

	writeCounter int64
	readCounter  int64
	readTimeout  time.Duration

	receiveBuffer bytes.Buffer
	sendBuffer    bytes.Buffer
}

// NewWebSocketConn returns a new WebSocketConn.
func NewWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	return &WebSocketConn{
		conn: conn,
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

	// flush buffer
	return c.flush()
}

// BufferedSend is currently not supported and falls back to normal Send.
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
	// reset and eventually grow buffer
	packetLength := pkt.Len()
	c.sendBuffer.Reset()
	c.sendBuffer.Grow(packetLength)
	buf := c.sendBuffer.Bytes()[0:packetLength]

	// encode packet
	n, err := pkt.Encode(buf)
	if err != nil {
		c.end(websocket.CloseInternalServerErr)
		return newTransportError(EncodeError, err)
	}

	// write packet to connection
	err = c.conn.BufferMessage(websocket.BinaryMessage, buf)
	if err != nil {
		c.end(websocket.CloseInternalServerErr)
		return newTransportError(NetworkError, err)
	}

	// increment write counter
	atomic.AddInt64(&c.writeCounter, int64(n))

	return nil
}

func (c *WebSocketConn) flush() error {
	err := c.conn.FlushMessages()
	if err != nil {
		c.conn.Close()
		return newTransportError(NetworkError, err)
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

// this implementation reuses the buffer in contrast to the official ReadMessage
// function
func (c *WebSocketConn) customReadMessage() ([]byte, error) {
	_, r, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}

	// reset and eventually grow buffer
	c.receiveBuffer.Reset()
	c.receiveBuffer.ReadFrom(r)

	return c.receiveBuffer.Bytes(), err
}

// Receive will read from the underlying connection and return a fully read
// packet. It will return an Error if there was an error while decoding or
// reading from the underlying connection.
//
// Note: Only one goroutine can Receive at the same time.
func (c *WebSocketConn) Receive() (packet.Packet, error) {
	c.rMutex.Lock()
	defer c.rMutex.Unlock()

	// read next message from connection
	buf, err := c.customReadMessage()

	// return ErrExpectedClose instead
	if _, ok := err.(*websocket.CloseError); ok {
		// ensure that the connection gets fully closed
		c.conn.Close()

		// the closing is always expected from the receiver side as it its
		// transmitted cleanly
		return nil, newTransportError(ConnectionClose, err)
	}

	// return read limit error instead
	if err == websocket.ErrReadLimit {
		c.end(websocket.CloseMessageTooBig)
		return nil, newTransportError(NetworkError, ErrReadLimitExceeded)
	}

	// return read timeout err instead
	if err != nil && strings.Contains(err.Error(), "i/o timeout") {
		c.end(websocket.CloseGoingAway)
		return nil, newTransportError(NetworkError, ErrReadTimeout)
	}

	// return on any other errors
	if err != nil {
		c.end(websocket.CloseAbnormalClosure)
		return nil, newTransportError(NetworkError, err)
	}

	// detect packet
	packetLength, t := packet.DetectPacket(buf)
	if packetLength == 0 {
		c.end(websocket.CloseInvalidFramePayloadData)
		return nil, newTransportError(DetectionError, ErrDetectionOverflow)
	}

	// allocate packet
	pkt, err := t.New()
	if err != nil {
		c.end(websocket.CloseInvalidFramePayloadData)
		return nil, newTransportError(DetectionError, err)
	}

	// decode packet
	n, err := pkt.Decode(buf)
	if err != nil {
		c.end(websocket.CloseInvalidFramePayloadData)
		return nil, newTransportError(DecodeError, err)
	}

	// increment read counter
	atomic.AddInt64(&c.readCounter, int64(n))

	// reset timeout
	c.resetTimeout()

	return pkt, nil
}

func (c *WebSocketConn) end(closeCode int) error {
	msg := websocket.FormatCloseMessage(closeCode, "")

	err := c.conn.WriteMessage(websocket.CloseMessage, msg)
	if err != nil {
		return newTransportError(NetworkError, err)
	}

	// ignore error as it would be raised by WriteMessage anyway
	c.conn.Close()

	return nil
}

// Close will close the underlying connection and cleanup resources. It will
// return an Error if there was an error while closing the underlying
// connection.
func (c *WebSocketConn) Close() error {
	return c.end(websocket.CloseNormalClosure)
}

// BytesWritten will return the number of bytes successfully written to
// the underlying connection.
func (c *WebSocketConn) BytesWritten() int64 {
	return c.writeCounter
}

// BytesRead will return the number of bytes successfully read from the
// underlying connection.
func (c *WebSocketConn) BytesRead() int64 {
	return c.readCounter
}

// SetReadLimit sets the maximum size of a packet that can be received.
// If the limit is greater than zero, Receive will close the connection and
// return an Error if receiving the next packet will exceed the limit.
func (c *WebSocketConn) SetReadLimit(limit int64) {
	c.conn.SetReadLimit(limit)
}

// SetReadTimeout sets the maximum time that can pass between reads.
// If no data is received in the set duration the connection will be closed
// and Read returns an error.
func (c *WebSocketConn) SetReadTimeout(timeout time.Duration) {
	c.readTimeout = timeout
	c.resetTimeout()
}

func (c *WebSocketConn) resetTimeout() {
	if c.readTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	} else {
		c.conn.SetReadDeadline(time.Time{})
	}
}
