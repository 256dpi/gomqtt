# gomqtt/router

[![Build Status](https://travis-ci.org/gomqtt/router.svg?branch=master)](https://travis-ci.org/gomqtt/router)
[![Coverage Status](https://coveralls.io/repos/github/gomqtt/router/badge.svg?branch=master)](https://coveralls.io/github/gomqtt/router?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/router?status.svg)](http://godoc.org/github.com/gomqtt/router)
[![Release](https://img.shields.io/github/release/gomqtt/router.svg)](https://github.com/gomqtt/router/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomqtt/router)](https://goreportcard.com/report/github.com/gomqtt/router)

**This go package implements a basic [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) application router component.**

_This package is WIP and may introduce breaking changes._

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/router
```

## Usage

```go
r := New()

done := make(chan struct{})

r.Handle("device/+id/#sensor", func(msg *packet.Message, params map[string]string) {
    fmt.Println(params["id"])
    fmt.Println(params["sensor"])
    fmt.Println(string(msg.Payload))

    close(done)
})

r.Start(client.NewOptions("mqtt://try:try@broker.shiftr.io"))

time.Sleep(2 * time.Second)

r.Publish("device/foo/bar/baz", []byte("42"), 0, false)

<-done

r.Stop()

// Output:
// foo
// bar/baz
// 42
```
