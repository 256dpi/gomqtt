package transport

import (
	"crypto/tls"
	"net"
)

// A NetServer accepts net.Conn based connections.
type NetServer struct {
	listener net.Listener
}

// NewNetServer creates a new TCP server that listens on the provided address.
func NewNetServer(address string) (*NetServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &NetServer{
		listener: listener,
	}, nil
}

// NewSecureNetServer creates a new TLS server that listens on the provided address.
func NewSecureNetServer(address string, config *tls.Config) (*NetServer, error) {
	listener, err := tls.Listen("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return &NetServer{
		listener: listener,
	}, nil
}

// Accept will return the next available connection or block until a
// connection becomes available, otherwise returns an Error.
func (s *NetServer) Accept() (Conn, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewNetConn(conn), nil
}

// Close will close the underlying listener and cleanup resources. It will
// return an Error if the underlying listener didn't close cleanly.
func (s *NetServer) Close() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

// Addr returns the server's network address.
func (s *NetServer) Addr() net.Addr {
	return s.listener.Addr()
}
