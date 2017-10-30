// Package transport implements functionality for handling MQTT 3.1.1
// (http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) connections.
package transport

import "errors"

// ErrUnsupportedProtocol is returned if either the launcher or dialer
// couldn't infer the protocol from the URL.
var ErrUnsupportedProtocol = errors.New("unsupported protocol")

// ErrAcceptAfterClose can be returned by a WebSocketServer during Accept()
// if the server has been already closed and the internal goroutine is dying.
//
// Note: this error is wrapped in an Error with NetworkError code.
var ErrAcceptAfterClose = errors.New("accept after close")
