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
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/tomb.v2"
)

var errManualClose = errors.New("internal: manual close")

// The WebSocketServer accepts websocket.Conn based connections.
type WebSocketServer struct {
	listener      net.Listener
	mux           *http.ServeMux
	fallback      http.Handler
	upgrader      *websocket.Upgrader
	incoming      chan *WebSocketConn
	originChecker func(r *http.Request) bool

	tomb tomb.Tomb
}

func newWebSocketServer(listener net.Listener) *WebSocketServer {
	ws := &WebSocketServer{
		listener: listener,
		upgrader: &websocket.Upgrader{
			HandshakeTimeout: 60 * time.Second,
			Subprotocols:     []string{"mqtt", "mqttv3.1"},
		},
		incoming: make(chan *WebSocketConn),
	}

	// add check origin method that uses the optional check origin function
	ws.upgrader.CheckOrigin = func(r *http.Request) bool {
		if ws.originChecker != nil {
			return ws.originChecker(r)
		}

		return true
	}

	return ws
}

// NewWebSocketServer creates a new WS server that listens on the provided address.
func NewWebSocketServer(address string) (*WebSocketServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	s := newWebSocketServer(listener)
	s.serveHTTP()

	return s, nil
}

// NewSecureWebSocketServer creates a new WSS server that listens on the
// provided address.
func NewSecureWebSocketServer(address string, config *tls.Config) (*WebSocketServer, error) {
	listener, err := tls.Listen("tcp", address, config)
	if err != nil {
		return nil, err
	}

	s := newWebSocketServer(listener)
	s.serveHTTP()

	return s, nil
}

func (s *WebSocketServer) serveHTTP() {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.requestHandler)

	h := &http.Server{
		Handler: s.mux,
	}

	s.tomb.Go(func() error {
		err := h.Serve(s.listener)

		// Server will always return an error
		return err
	})
}

// SetFallback will register a http.Handler that gets called if a request is not
// a WebSocket upgrade request.
func (s *WebSocketServer) SetFallback(handler http.Handler) {
	s.fallback = handler
}

// SetOriginChecker sets an optional function that allows check the request origin
// before accepting the connection.
func (s *WebSocketServer) SetOriginChecker(fn func(r *http.Request) bool) {
	s.originChecker = fn
}

func (s *WebSocketServer) requestHandler(w http.ResponseWriter, r *http.Request) {
	// run fallback if request is not an upgrade
	if r.Header.Get("Upgrade") != "websocket" && s.fallback != nil {
		s.fallback.ServeHTTP(w, r)
		return
	}

	// run WebSocket upgrader
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// upgrader already responded to request
		return
	}

	// create connection
	webSocketConn := NewWebSocketConn(conn)

	select {
	case s.incoming <- webSocketConn:
	case <-s.tomb.Dying():
		webSocketConn.Close()
	}
}

// Accept will return the next available connection or block until a
// connection becomes available, otherwise returns an Error.
func (s *WebSocketServer) Accept() (Conn, error) {
	select {
	case <-s.tomb.Dying():
		if s.tomb.Err() == errManualClose {
			// server has been closed manually
			return nil, ErrAcceptAfterClose
		}

		// return the previously caught error
		return nil, s.tomb.Err()
	case conn := <-s.incoming:
		return conn, nil
	}
}

// Close will close the underlying listener and cleanup resources. It will
// return an Error if the underlying listener didn't close cleanly.
func (s *WebSocketServer) Close() error {
	s.tomb.Kill(errManualClose)

	err := s.listener.Close()
	s.tomb.Wait()

	if err != nil {
		return err
	}

	return nil
}

// Addr returns the server's network address.
func (s *WebSocketServer) Addr() net.Addr {
	return s.listener.Addr()
}
