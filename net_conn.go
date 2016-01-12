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
	"io"
	"net"
	"sync/atomic"
	"errors"

	"github.com/gomqtt/packet"
)

var ErrDetectionOverflow = errors.New("detection length overflow (>5)")
var ErrReadLimitExceeded = errors.New("read limit exceeded")

// The NetConn wraps a TCP based connection.
type NetConn struct {
	conn net.Conn
	reader *bufio.Reader

	writeCounter int64
	readCounter int64
	readLimit int64
}

// NewNetConn returns a new NetConn.
func NewNetConn(conn net.Conn) *NetConn {
	return &NetConn{
		conn: conn,
		reader: bufio.NewReader(conn),
	}
}

func (c *NetConn) Send(pkt packet.Packet) error {
	// allocate buffer
	buf := make([]byte, pkt.Len())

	// encode packet
	_, err := pkt.Encode(buf)
	if err != nil {
		c.conn.Close()
		return newTransportError(EncodeError, err)
	}

	// write buffer to connection
	bytesWritten, err := c.conn.Write(buf)
	if err != nil {
		c.conn.Close()
		return newTransportError(ConnectionError, err)
	}

	// increment write counter
	atomic.AddInt64(&c.writeCounter, int64(bytesWritten))

	return nil
}

func (c *NetConn) Receive() (packet.Packet, error) {
	// initial detection length
	detectionLength := 2

	for {
		// check length
		if detectionLength > 5 {
			c.conn.Close()
			return nil, newTransportError(DetectionError, ErrDetectionOverflow)
		}

		// try read detection bytes
		header, err := c.reader.Peek(detectionLength)
		if err == io.EOF {
			c.conn.Close()
			return nil, newTransportError(ExpectedClose, err)
		} else if err != nil {
			c.conn.Close()
			return nil, newTransportError(ConnectionError, err)
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
			c.conn.Close()
			return nil, newTransportError(ReadLimitExceeded, ErrReadLimitExceeded)
		}

		// create packet
		pkt, err := packetType.New()
		if err != nil {
			c.conn.Close()
			return nil, newTransportError(DetectionError, err)
		}

		// allocate buffer
		buf := make([]byte, packetLength)

		// read whole packet
		bytesRead, err := io.ReadFull(c.reader, buf)
		if err == io.EOF {
			c.conn.Close()
			return nil, newTransportError(ExpectedClose, err)
		} else if err != nil {
			c.conn.Close()
			return nil, newTransportError(ConnectionError, err)
		}

		// decode buffer
		_, err = pkt.Decode(buf)
		if err != nil {
			c.conn.Close()
			return nil, newTransportError(DecodeError, err)
		}

		// increment counter
		atomic.AddInt64(&c.readCounter, int64(bytesRead))

		return pkt, nil
	}
}

func (c *NetConn) Close() error {
	err := c.conn.Close()
	if err != nil {
		return newTransportError(ConnectionError, err)
	}

	return nil
}

func (c *NetConn) BytesWritten() int64 {
	return c.writeCounter
}

func (c *NetConn) BytesRead() int64 {
	return c.readCounter
}

func (c *NetConn) SetReadLimit(limit int64) {
	c.readLimit = limit
}
