# gomqtt/packet

[![Build Status](https://travis-ci.org/gomqtt/packet.svg?branch=master)](https://travis-ci.org/gomqtt/packet)
[![Coverage Status](https://coveralls.io/repos/github/gomqtt/packet/badge.svg?branch=master)](https://coveralls.io/github/gomqtt/packet?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/packet?status.svg)](http://godoc.org/github.com/gomqtt/packet)
[![Release](https://img.shields.io/github/release/gomqtt/packet.svg)](https://github.com/gomqtt/packet/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomqtt/packet)](https://goreportcard.com/report/github.com/gomqtt/packet)

**Package packet implements functionality for encoding and decoding [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) packets.**

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/packet
```

## Usage

```go
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
pkt2, err := mt.New();
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
```

More details can be found in the [documentation](http://godoc.org/github.com/gomqtt/packet).

## Credits

This package has been originally extracted and contributed by @zhenjl from the
[surgemq](https://github.com/surgemq/surgemq) project.
