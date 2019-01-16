package transport

import (
	"net"
	"time"
)

// A NetConn is a wrapper around a basic TCP connection.
type NetConn struct {
	*BaseConn

	conn net.Conn
}

// NewNetConn returns a new NetConn.
func NewNetConn(conn net.Conn, maxWriteDelay time.Duration) *NetConn {
	return &NetConn{
		BaseConn: NewBaseConn(conn, maxWriteDelay),
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

// UnderlyingConn returns the underlying net.Conn.
func (c *NetConn) UnderlyingConn() net.Conn {
	return c.conn
}
