# gomqtt/topic

[![Circle CI](https://img.shields.io/circleci/project/gomqtt/topic.svg)](https://circleci.com/gh/gomqtt/topic)
[![Coverage Status](https://coveralls.io/repos/gomqtt/topic/badge.svg?branch=master&service=github)](https://coveralls.io/github/gomqtt/topic?branch=master)
[![GoDoc](https://godoc.org/github.com/gomqtt/topic?status.svg)](http://godoc.org/github.com/gomqtt/topic)
[![Release](https://img.shields.io/github/release/gomqtt/topic.svg)](https://github.com/gomqtt/topic/releases)

**Package topic implements functionality for working with [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) topics.**

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/topic
```

## Usage

```go
tree := NewTree()

tree.Add("foo/bar", 1)
tree.Add("foo/bar/baz", 2)
tree.Add("foo/+", 3)
tree.Add("foo/#", 4)

fmt.Println(tree.Match("foo/bar"))
fmt.Println(tree.Match("foo/bar/baz"))

// Output:
// [4 3 1]
// [4 2]
```
