# gomqtt/broker

[![Build Status](https://travis-ci.org/gomqtt/broker.svg?branch=master)](https://travis-ci.org/gomqtt/broker)
[![Coverage Status](https://coveralls.io/repos/github/gomqtt/broker/badge.svg?branch=master)](https://coveralls.io/github/gomqtt/broker?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/broker?status.svg)](http://godoc.org/github.com/gomqtt/broker)
[![Release](https://img.shields.io/github/release/gomqtt/broker.svg)](https://github.com/gomqtt/broker/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomqtt/broker)](https://goreportcard.com/report/github.com/gomqtt/broker)

**Package broker provides an extensible [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) broker implementation.**

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/broker
```

## Usage

```go
server, err := transport.Launch("tcp://localhost:8080")
if err != nil {
    panic(err)
}

engine := NewEngine()
engine.Accept(server)

c := client.New()
wait := make(chan struct{})

c.Callback = func(msg *packet.Message, err error) {
    if err != nil {
        panic(err)
    }

    fmt.Println(msg.String())
    close(wait)
}

cf, err := c.Connect(client.NewOptions("tcp://localhost:8080"))
if err != nil {
    panic(err)
}

cf.Wait()

sf, err := c.Subscribe("test", 0)
if err != nil {
    panic(err)
}

sf.Wait()

pf, err := c.Publish("test", []byte("test"), 0, false)
if err != nil {
    panic(err)
}

pf.Wait()

<-wait

err = c.Disconnect()
if err != nil {
    panic(err)
}

err = server.Close()
if err != nil {
    panic(err)
}

engine.Close()

// Output:
// <Message Topic="test" QOS=0 Retain=false Payload=[116 101 115 116]>
```
