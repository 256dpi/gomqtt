# gomqtt/packet

[![Circle CI](https://img.shields.io/circleci/project/gomqtt/packet.svg)](https://circleci.com/gh/gomqtt/packet)
[![Coverage Status](https://coveralls.io/repos/gomqtt/packet/badge.svg?branch=master&service=github)](https://coveralls.io/github/gomqtt/packet?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/packet?status.svg)](http://godoc.org/github.com/gomqtt/packet)
[![Release](https://img.shields.io/github/release/gomqtt/packet.svg)](https://github.com/gomqtt/packet/releases)
[![Go Report Card](http://goreportcard.com/badge/gomqtt/packet)](http://goreportcard.com/report/gomqtt/packet)

**Package packet implements functionality for encoding and decoding [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) packets.**

## Features

- Built around plain byte slices.
- Zero allocation encoding and decoding.
- Builtin packet detection.
- No overflows thanks to input fuzzing.
- Full test coverage.

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/packet
```

## Usage

Create a new packet and encode it:

```go
// Create new packet.
pkt1 := NewConnectPacket()
pkt1.Username = []byte("gomqtt")
pkt1.Password = []byte("amazing!")

// Allocate buffer.
buf := make([]byte, pkt1.Len())

// Encode the packet.
if _, err := pkt1.Encode(buf); err != nil {
    // there was an error while encoding
    panic(err)
}

// Send buffer off the wire.
```

Decode bytes to a packet:

```go
// Get buffer from the wire.

// Detect packet.
l, t := DetectPacket(buf)

// Check length
if l == 0 {
    // buffer not complete yet
    return
}

// Create packet.
pkt2, err := t.New()
if err != nil {
    // type is invalid
    panic(err)
}

// Decode packet.
_, err = pkt2.Decode(buf)
if err != nil {
    // there was an error while decoding
    panic(err)
}

switch pkt2.Type() {
case CONNECT:
    c := pkt2.(*ConnectPacket)
    fmt.Println(string(c.Username))
    fmt.Println(string(c.Password))
}

// Output:
// gomqtt
// amazing!
```

More details can be found in the [documentation](http://godoc.org/github.com/gomqtt/packet).

## Credits

This package has been originally extracted and contributed by @zhenjl from the
[surgemq](https://github.com/surgemq/surgemq) project.
