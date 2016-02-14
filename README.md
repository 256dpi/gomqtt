# gomqtt/transport

[![Circle CI](https://img.shields.io/circleci/project/gomqtt/transport.svg)](https://circleci.com/gh/gomqtt/transport)
[![Coverage Status](https://coveralls.io/repos/gomqtt/transport/badge.svg?branch=master&service=github)](https://coveralls.io/github/gomqtt/transport?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/transport?status.svg)](http://godoc.org/github.com/gomqtt/transport)
[![Release](https://img.shields.io/github/release/gomqtt/transport.svg)](https://github.com/gomqtt/transport/releases)
[![Go Report Card](http://goreportcard.com/badge/gomqtt/transport)](http://goreportcard.com/report/gomqtt/transport)

**Package transport implements functionality for handling [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) connections.**

## Features

- Support for `net.Conn` based connections.
- Support for `websocket.Conn` based connections through <https://github.com/gorilla/websocket>.
- Shorthand `Dial` function for initiating connections.
- Shorthand `Launch` function for launching servers.
- Rich `Error` interface for enhanced error handling.

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/transport
```

## Usage

```go
// launch server
server, err := Launch("tcp://localhost:1883")
if err != nil {
    panic(err)
}

go func(){
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
conn, err := Dial("tcp://localhost:1883")
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
```
