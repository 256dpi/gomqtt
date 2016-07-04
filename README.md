# gomqtt/spec

[![Circle CI](https://img.shields.io/circleci/project/gomqtt/spec.svg)](https://circleci.com/gh/gomqtt/spec)
[![Coverage Status](https://coveralls.io/repos/gomqtt/spec/badge.svg?branch=master&service=github)](https://coveralls.io/github/gomqtt/spec?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/spec?status.svg)](http://godoc.org/github.com/gomqtt/spec)
[![Release](https://img.shields.io/github/release/gomqtt/spec.svg)](https://github.com/gomqtt/spec/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomqtt/spec)](http://goreportcard.com/report/gomqtt/spec)

**Package spec provides functionality for testing [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) broker implementations.**

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/spec
```

## Usage

```go
func TestSpec(t *testing.T) {
	config := AllFeatures()
	config.URL = "tcp://localhost:1883"

	// mosquitto specific config
	config.Authentication = false
	config.MessageRetainWait = 300 * time.Millisecond
	config.NoMessageWait = 100 * time.Millisecond

	Run(t, config)
}
```
