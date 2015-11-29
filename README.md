# gomqtt/message

[![Circle CI](https://img.shields.io/circleci/project/gomqtt/message.svg)](https://circleci.com/gh/gomqtt/message)
[![Coverage Status](https://coveralls.io/repos/gomqtt/message/badge.svg?branch=master&service=github)](https://coveralls.io/github/gomqtt/message?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/message?status.svg)](http://godoc.org/github.com/gomqtt/message)
[![Release](https://img.shields.io/github/release/gomqtt/message.svg)](https://github.com/gomqtt/message/releases)

**This go package implements functionality for encoding and decoding [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) messages.**

## Features

- Zero allocation encoding/decoding.
- Builtin message detection.
- No overflows thanks to input fuzzing.

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/message
```

## Usage

Create a new message and encode it:

```go
// Create new message.
msg1 := NewConnectMessage()
msg1.Username = []byte("gomqtt")
msg1.Password = []byte("amazing!")

// Allocate buffer.
buf := make([]byte, msg1.Len())

// Encode the message.
if _, err := msg1.Encode(buf); err != nil {
    // there was an error while encoding
    panic(err)
}
```

Decode bytes to a message:

```go
// Detect message.
l, mt := DetectMessage(buf)

// Check length
if l == 0 {
    // buffer not complete yet
    return
}

// Create message.
msg2, err := mt.New()
if err != nil {
    // message type is invalid
    panic(err)
}

// Decode message.
_, err = msg2.Decode(buf)
if err != nil {
    // there was an error while decoding
    panic(err)
}
```

More details can be found in the [documentation](http://godoc.org/github.com/gomqtt/message).

## Credits

This package has been originally extracted and contributed by @zhenjl from the
[surgemq](https://github.com/surgemq/surgemq) project.
