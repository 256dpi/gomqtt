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
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gorilla/websocket"
)

type webSocketStream struct {
	conn   *websocket.Conn
	reader io.Reader
}

func (s *webSocketStream) Read(p []byte) (int, error) {
	for {
		// get next reader
		if s.reader == nil {
			// TODO: return error if message is not binary.
			_, reader, err := s.conn.NextReader()
			if _, ok := err.(*websocket.CloseError); ok {
				// convert CloseError to EOF
				return 0, io.EOF
			} else if err != nil {
				return 0, err
			}

			// set current reader
			s.reader = reader
		}

		// read data
		n, err := s.reader.Read(p)
		if err == io.EOF {
			// move on to next message on EOF
			s.reader = nil
			continue
		}

		return n, err
	}
}

// The WebSocketConn wraps a gorilla WebSocket.Conn. The implementation supports
// packets that are chunked over several WebSocket messages and packets that are
// coalesced to one WebSocket message.
type WebSocketConn struct {
	conn   *websocket.Conn
	reader *bufio.Reader
	writer io.WriteCloser

	flushTimer *time.Timer
	flushError error

	sMutex sync.Mutex
	rMutex sync.Mutex

	writeCounter int64
	readCounter  int64
	readLimit    int64
	readTimeout  time.Duration

	receiveBuffer bytes.Buffer
	sendBuffer    bytes.Buffer
}

// NewWebSocketConn returns a new WebSocketConn.
func NewWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	return &WebSocketConn{
		conn: conn,
		reader: bufio.NewReader(&webSocketStream{
			conn: conn,
		}),
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

	// create writer if missing
	if c.writer == nil {
		writer, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			c.end(websocket.CloseInternalServerErr)
			return newTransportError(NetworkError, err)
		}

		c.writer = writer
	}

	// write packet to writer
	_, err = c.writer.Write(buf)
	if err != nil {
		c.end(websocket.CloseInternalServerErr)
		return newTransportError(NetworkError, err)
	}

	// increment write counter
	atomic.AddInt64(&c.writeCounter, int64(n))

	return nil
}

func (c *WebSocketConn) flush() error {
	// close writer if available
	if c.writer != nil {
		err := c.writer.Close()
		if err != nil {
			c.conn.Close()
			return newTransportError(NetworkError, err)
		}

		c.writer = nil
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

	// initial detection length
	detectionLength := 2

	for {
		// check length
		if detectionLength > 5 {
			c.end(websocket.CloseInvalidFramePayloadData)
			return nil, newTransportError(DetectionError, ErrDetectionOverflow)
		}

		// try read detection bytes
		header, err := c.reader.Peek(detectionLength)
		if err == io.EOF && len(header) == 0 {
			// only if Peek returned no bytes the close is expected
			c.conn.Close()
			return nil, newTransportError(ConnectionClose, err)
		} else if err != nil && strings.Contains(err.Error(), "i/o timeout") {
			// the read timed out
			c.end(websocket.CloseGoingAway)
			return nil, newTransportError(NetworkError, ErrReadTimeout)
		} else if err != nil {
			c.end(websocket.CloseAbnormalClosure)
			return nil, newTransportError(NetworkError, err)
		}

		// detect packet
		packetLength, packetType := packet.DetectPacket(header)

		// on zero packet length:
		// increment detection length and try again
		if packetLength <= 0 {
			detectionLength++
			continue
		}

		// check read limit
		if c.readLimit > 0 && int64(packetLength) > c.readLimit {
			c.end(websocket.CloseMessageTooBig)
			return nil, newTransportError(NetworkError, ErrReadLimitExceeded)
		}

		// create packet
		pkt, err := packetType.New()
		if err != nil {
			c.end(websocket.CloseInvalidFramePayloadData)
			return nil, newTransportError(DetectionError, err)
		}

		// reset and eventually grow buffer
		c.receiveBuffer.Reset()
		c.receiveBuffer.Grow(packetLength)
		buf := c.receiveBuffer.Bytes()[0:packetLength]

		// read whole packet
		bytesRead, err := io.ReadFull(c.reader, buf)
		if err != nil && strings.Contains(err.Error(), "i/o timeout") {
			// the read timed out
			c.end(websocket.CloseGoingAway)
			return nil, newTransportError(NetworkError, ErrReadTimeout)
		} else if err != nil {
			c.end(websocket.CloseAbnormalClosure)

			// even if EOF is returned we consider it an network error
			return nil, newTransportError(NetworkError, err)
		}

		// decode buffer
		_, err = pkt.Decode(buf)
		if err != nil {
			c.end(websocket.CloseInvalidFramePayloadData)
			return nil, newTransportError(DecodeError, err)
		}

		// increment counter
		atomic.AddInt64(&c.readCounter, int64(bytesRead))

		// reset timeout
		c.resetTimeout()

		return pkt, nil
	}
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
	c.sMutex.Lock()
	defer c.sMutex.Unlock()

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
	c.readLimit = limit
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
