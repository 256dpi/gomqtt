# gomqtt/router

[![Build Status](https://travis-ci.org/gomqtt/router.svg?branch=master)](https://travis-ci.org/gomqtt/router)
[![Coverage Status](https://coveralls.io/repos/github/gomqtt/router/badge.svg?branch=master)](https://coveralls.io/github/gomqtt/router?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/router?status.svg)](http://godoc.org/github.com/gomqtt/router)
[![Release](https://img.shields.io/github/release/gomqtt/router.svg)](https://github.com/gomqtt/router/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomqtt/router)](https://goreportcard.com/report/github.com/gomqtt/router)

**This go package implements a basic [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) application router component.**

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/router
```

## Usage

```go
r := New()

done := make(chan struct{})

r.Handle("device/+id/#sensor", Logger(func(w ResponseWriter, r *Request) {
    w.Publish(&packet.Message{
        Topic:   "finish/data",
        Payload: []byte("7"),
    })
}))

r.Handle("finish/data", Logger(func(w ResponseWriter, r *Request) {
    close(done)
}))

config := client.NewConfig("mqtt://try:try@broker.shiftr.io")

r.Start(config)

time.Sleep(2 * time.Second)

r.Publish(&packet.Message{
    Topic:   "device/foo/bar/baz",
    Payload: []byte("42"),
})

<-done

r.Stop()

// Output:
// New Request: &{<Message Topic="device/foo/bar/baz" QOS=0 Retain=false Payload=[52 50]> map[id:foo sensor:bar/baz]}
// Publishing: <Message Topic="finish/data" QOS=0 Retain=false Payload=[55]>
// New Request: &{<Message Topic="finish/data" QOS=0 Retain=false Payload=[55]> map[]}
```
