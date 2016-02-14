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
	"fmt"

	"github.com/gomqtt/packet"
)

func Example() {
	// launch server
	server, err := Launch("tcp://localhost:1337")
	if err != nil {
		panic(err)
	}

	go func() {
		// accept next incoming connection
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}

		// receive next packet
		pkt, err := conn.Receive()
		if err != nil {
			panic(err)
		}

		// check packet type
		if _, ok := pkt.(*packet.ConnectPacket); ok {
			// send a connack packet
			err = conn.Send(packet.NewConnackPacket())
			if err != nil {
				panic(err)
			}
		} else {
			panic("unexpected packet")
		}
	}()

	// dial to server
	conn, err := Dial("tcp://localhost:1337")
	if err != nil {
		panic(err)
	}

	// send connect packet
	err = conn.Send(packet.NewConnectPacket())
	if err != nil {
		panic(err)
	}

	// receive next packet
	pkt, err := conn.Receive()
	if err != nil {
		panic(err)
	}

	// check packet type
	if connackPacket, ok := pkt.(*packet.ConnackPacket); ok {
		fmt.Println(connackPacket)

		// close connection
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	} else {
		panic("unexpected packet")
	}

	// close server
	err = server.Close()
	if err != nil {
		panic(err)
	}

	// Output:
	// <ConnackPacket SessionPresent=false ReturnCode=0>
}
