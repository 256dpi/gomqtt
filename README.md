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

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/transport
```
