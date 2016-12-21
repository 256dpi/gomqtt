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

import "net"

// A Server is a local port on which incoming connections can be accepted.
type Server interface {
	// Accept will return the next available connection or block until a
	// connection becomes available, otherwise returns an Error.
	Accept() (Conn, error)

	// Close will close the underlying listener and cleanup resources. It will
	// return an Error if the underlying listener didn't close cleanly.
	Close() error

	// Addr returns the server's network address.
	Addr() net.Addr
}
