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
	"fmt"
)

// ErrUnsupportedProtocol is returned if either the launcher or dialer
// couldn't infer the protocol from the URL.
//
// Note: this error is wrapped in an Error with a LaunchError or DialError code.
var ErrUnsupportedProtocol = errors.New("unsupported protocol")

// ErrDetectionOverflow can be returned during a Receive if the next packet
// couldn't be detect from the initial header bytes.
//
// Note: this error is wrapped in an Error with a DetectionError code.
var ErrDetectionOverflow = errors.New("detection length overflow (>5)")

// ErrReadLimitExceeded can be returned during a Receive if the connection
// exceeded its read limit.
//
// Note: this error is wrapped in an Error with a NetworkError code.
var ErrReadLimitExceeded = errors.New("read limit exceeded")

// ErrReadTimeout can be returned by Receive if the connection did not Read
// data in the by SetReadTimeout specified duration.
var ErrReadTimeout = errors.New("read timeout")

// ErrAcceptAfterClose can be returned by a WebSocketServer during Accept()
// if the server has been already closed and the internal goroutine is dying.
//
// Note: this error is wrapped in an Error with NetworkError code.
var ErrAcceptAfterClose = errors.New("accept after close")

// The ErrorCode provides a context to various errors.
type ErrorCode int

const (
	_ ErrorCode = iota

	// ConnectionClose marks errors that symbolizes the close of a connection.
	ConnectionClose

	// DialError marks errors that came up during a Dial call.
	DialError

	// LaunchError marks errors that came up during a Launch call.
	LaunchError

	// EncodeError marks errors that came up within Send and are returned by the
	// packets Encode functions.
	EncodeError

	// DecodeError marks errors that came up within Receive and are returned by
	// the packets Decode functions.
	DecodeError

	// DetectionError marks errors that cam up within Receive.
	DetectionError

	// NetworkError marks errors that are returned by the underlying connection.
	NetworkError
)

// An Error a wraps standard errors and provides additional context information.
type Error interface {
	error

	// Code will return the corresponding error code.
	Code() ErrorCode

	// Err will return the original error.
	Err() error
}

// IsConnectionCloseError returns true if the supplied error is a ConnectionClose
// transport error.
func IsConnectionCloseError(value interface{}) bool {
	if err, ok := value.(*transportError); ok {
		return err.code == ConnectionClose
	}

	return false
}

type transportError struct {
	code ErrorCode
	err  error
}

func newTransportError(code ErrorCode, err error) *transportError {
	return &transportError{
		code: code,
		err:  err,
	}
}

func (err *transportError) Error() string {
	switch err.code {
	case ConnectionClose:
		return fmt.Sprintf("connection close: %s", err.err.Error())
	case DialError:
		return fmt.Sprintf("dial error: %s", err.err.Error())
	case LaunchError:
		return fmt.Sprintf("launch error: %s", err.err.Error())
	case EncodeError:
		return fmt.Sprintf("encode error: %s", err.err.Error())
	case DecodeError:
		return fmt.Sprintf("decode error: %s", err.err.Error())
	case DetectionError:
		return fmt.Sprintf("detection error: %s", err.err.Error())
	case NetworkError:
		return fmt.Sprintf("network error: %s", err.err.Error())
	}

	return fmt.Sprintf("unknown error: %s", err.err.Error())
}

func (err *transportError) Code() ErrorCode {
	return err.code
}

func (err *transportError) Err() error {
	return err.err
}
