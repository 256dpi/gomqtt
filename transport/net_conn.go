package transport

import (
	"net"
)

// A NetConn is a wrapper around a basic TCP connection.
type NetConn struct {
	*BaseConn

	conn net.Conn
}

// NewNetConn returns a new NetConn.
func NewNetConn(conn net.Conn) *NetConn {
	return &NetConn{
		BaseConn: NewBaseConn(conn),
		conn:     conn,
	}
}

// LocalAddr returns the local network address.
func (c *NetConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *NetConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetBuffers allows to set the buffer sizes used by the operating system.
func (c *NetConn) SetBuffers(read, write int) {
	// set tcp conn
	if tcpConn, ok := c.conn.(*net.TCPConn); ok {
		tcpConn.SetReadBuffer(read)
		tcpConn.SetWriteBuffer(write)
	}
}

// UnderlyingConn returns the underlying net.Conn.
func (c *NetConn) UnderlyingConn() net.Conn {
	return c.conn
}
