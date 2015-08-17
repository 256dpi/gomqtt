# gomqtt/message

[![Circle CI](https://img.shields.io/circleci/project/gomqtt/message.svg)](https://circleci.com/gh/gomqtt/message)
[![GoDoc](https://godoc.org/github.com/gomqtt/message?status.svg)](http://godoc.org/github.com/gomqtt/message)
[![Release](https://img.shields.io/github/release/gomqtt/message.svg)](https://github.com/gomqtt/message/releases)

This go package is an encoder/decoder library for
[MQTT 3.1](http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html)
and [MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) messages.

## Installation

Get it using go's standard toolset:

```bash
$ go get github.com/gomqtt/message
```

## Usage

Create a new message and encode it:

```go
// Create new message.
msg := NewConnectMessage()

msg.SetWillQos(1)
msg.SetVersion(4)
msg.SetCleanSession(true)
msg.SetClientId([]byte("surgemq"))
msg.SetKeepAlive(10)
msg.SetWillTopic([]byte("will"))
msg.SetWillMessage([]byte("send me home"))
msg.SetUsername([]byte("surgemq"))
msg.SetPassword([]byte("verysecret"))

// Encode the message and get an io.Reader.
r, n, err := msg.Encode()
if err == nil {
    return err
}

// Write bytes into a connection.
m, err := io.CopyN(conn, r, int64(n))
if err != nil {
    return err
}

fmt.Printf("Sent %d bytes of %s message", m, msg.Name())
```

Receive bytes and decode them:

```go
// Create a buffered IO reader for a connection.
br := bufio.NewReader(conn)

// Peek at the first byte, which contains the message type.
b, err := br.Peek(1)
if err != nil {
    return err
}

// Extract the type from the first byte.
t := MessageType(b[0] >> 4)

// Create a new message
msg, err := t.New()
if err != nil {
    return err
}

// Decode it from the bufio.Reader.
n, err := msg.Decode(br)
if err != nil {
    return err
}
```

More details can be found in the [documentation](http://godoc.org/github.com/gomqtt/message).
