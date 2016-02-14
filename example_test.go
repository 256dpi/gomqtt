package packet

import "fmt"

func Example() {
	/* Packet Encoding */

	// Create new packet.
	pkt1 := NewConnectPacket()
	pkt1.Username = "gomqtt"
	pkt1.Password = "amazing!"

	// Allocate buffer.
	buf := make([]byte, pkt1.Len())

	// Encode the packet.
	if _, err := pkt1.Encode(buf); err != nil {
		panic(err) // error while encoding
	}

	/* Packet Decoding */

	// Detect packet.
	l, mt := DetectPacket(buf)

	// Check length
	if l == 0 {
		return // buffer not complete yet
	}

	// Create packet.
	pkt2, err := mt.New()
	if err != nil {
		panic(err) // packet type is invalid
	}

	// Decode packet.
	_, err = pkt2.Decode(buf)
	if err != nil {
		panic(err) // there was an error while decoding
	}

	switch pkt2.Type() {
	case CONNECT:
		c := pkt2.(*ConnectPacket)
		fmt.Println(c.Username)
		fmt.Println(c.Password)
	}

	// Output:
	// gomqtt
	// amazing!
}
