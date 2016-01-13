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
	"net"
	"crypto/tls"
)

type NetServer struct {
	listener net.Listener
}

func NewNetServer(address string) (*NetServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, newTransportError(NetworkError, err)
	}

	return &NetServer{
		listener: listener,
	}, nil
}

func NewSecureNetServer(address string, config *tls.Config) (*NetServer, error) {
	listener, err := tls.Listen("tcp", address, config)
	if err != nil {
		return nil, newTransportError(NetworkError, err)
	}

	return &NetServer{
		listener: listener,
	}, nil
}

func (s *NetServer) Accept() (Conn, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, newTransportError(NetworkError, err)
	}

	return NewNetConn(conn), nil
}

func (s *NetServer) Close() error {
	err := s.listener.Close()
	if err != nil {
		return newTransportError(NetworkError, err)
	}

	return nil
}
