package transport

import (
	"fmt"

	"github.com/256dpi/gomqtt/packet"
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
			err = conn.Send(packet.NewConnack())
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
	if connackPacket, ok := pkt.(*packet.Connack); ok {
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
	// <Connack SessionPresent=false ReturnCode=0>
}
