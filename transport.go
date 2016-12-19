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

// Package transport implements functionality for handling MQTT 3.1.1
// (http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) connections.
package transport

import "errors"

// ErrUnsupportedProtocol is returned if either the launcher or dialer
// couldn't infer the protocol from the URL.
var ErrUnsupportedProtocol = errors.New("unsupported protocol")

// ErrDetectionOverflow can be returned during a Receive if the next packet
// couldn't be detect from the initial header bytes.
//
// Note: this error is wrapped in an Error with a DetectionError code.
var ErrDetectionOverflow = errors.New("detection overflow")

// ErrReadLimitExceeded can be returned during a Receive if the connection
// exceeded its read limit.
//
// Note: this error is wrapped in an Error with a NetworkError code.
var ErrReadLimitExceeded = errors.New("read limit exceeded")

// ErrReadTimeout can be returned by Receive if the connection did not Read
// data in the by SetReadTimeout specified duration.
//
// Note: this error is wrapped in an Error with a NetworkError code.
var ErrReadTimeout = errors.New("read timeout")

// ErrAcceptAfterClose can be returned by a WebSocketServer during Accept()
// if the server has been already closed and the internal goroutine is dying.
//
// Note: this error is wrapped in an Error with NetworkError code.
var ErrAcceptAfterClose = errors.New("accept after close")
