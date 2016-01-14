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
	"sync/atomic"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
)

// The WebSocketConn wraps a gorilla WebSocket.Conn.
type WebSocketConn struct {
	conn *websocket.Conn

	writeCounter int64
	readCounter  int64
}

// NewWebSocketConn returns a new WebSocketConn.
func NewWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	return &WebSocketConn{
		conn: conn,
	}
}

// Send will write the packet to the underlying connection. It will return
// an Error if there was an error while encoding or writing to the
// underlying connection.
func (c *WebSocketConn) Send(pkt packet.Packet) error {
	// allocate buffer
	buf := make([]byte, pkt.Len())

	// encode packet
	n, err := pkt.Encode(buf)
	if err != nil {
		c.end()
		return newTransportError(EncodeError, err)
	}

	// write packet to connection
	err = c.conn.WriteMessage(websocket.BinaryMessage, buf)
	if err != nil {
		c.conn.Close()
		return newTransportError(NetworkError, err)
	}

	// increment write counter
	atomic.AddInt64(&c.writeCounter, int64(n))

	return nil
}

// Receive will read from the underlying connection and return a fully read
// packet. It will return an Error if there was an error while decoding or
// reading from the underlying connection.
func (c *WebSocketConn) Receive() (packet.Packet, error) {
	// read next message from connection
	_, buf, err := c.conn.ReadMessage()

	// return ErrExpectedClose instead
	if closeErr, ok := err.(*websocket.CloseError); ok {
		if closeErr.Code == websocket.CloseNormalClosure {
			c.conn.Close()
			return nil, newTransportError(ExpectedClose, err)
		} else if closeErr.Code == websocket.CloseMessageTooBig {
			// TODO: shouldn't be an expected close
			return nil, newTransportError(ExpectedClose, err)
		}
	}

	// return read limit error instead
	if err == websocket.ErrReadLimit {
		c.conn.Close()
		return nil, newTransportError(NetworkError, err)
	}

	// return on any other errors
	if err != nil {
		c.conn.Close()
		return nil, newTransportError(NetworkError, err)
	}

	// detect packet
	packetLength, t := packet.DetectPacket(buf)
	if packetLength == 0 {
		c.end()
		return nil, newTransportError(DetectionError, ErrDetectionOverflow)
	}

	// allocate packet
	pkt, err := t.New()
	if err != nil {
		c.end()
		return nil, newTransportError(DetectionError, err)
	}

	// decode packet
	n, err := pkt.Decode(buf)
	if err != nil {
		c.end()
		return nil, newTransportError(DecodeError, err)
	}

	// increment read counter
	atomic.AddInt64(&c.readCounter, int64(n))

	return pkt, nil
}

func (c *WebSocketConn) end() error {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")

	err := c.conn.WriteMessage(websocket.CloseMessage, msg)
	if err != nil {
		return newTransportError(NetworkError, err)
	}

	err = c.conn.Close()
	if err != nil {
		return newTransportError(NetworkError, err)
	}

	return nil
}

// Close will close the underlying connection and cleanup resources. It will
// return an Error if there was an error while closing the underlying
// connection.
func (c *WebSocketConn) Close() error {
	return c.end()
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
