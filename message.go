// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
This go package is an encoder/decoder library for
[MQTT 3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) messages.

	MQTT is a client server publish/subscribe messaging transport protocol. It is
	light weight, open, simple, and designed so as to be easy to implement. These
	characteristics make it ideal for use in many situations, including constrained
	environments such as for communication in Machine to Machine (M2M) and Internet
	of Things (IoT) contexts where a small code footprint is required and/or network
	bandwidth is at a premium.

	The MQTT protocol works by exchanging a series of MQTT messages in a defined way.
	The protocol runs over TCP/IP, or over other network protocols that provide
	ordered, lossless, bi-directional connections.


There are two main items to take note in this package. The first is:

	type MessageType byte

MessageType is the type representing the MQTT packet types. In the MQTT spec, MQTT
control packet type is represented as a 4-bit unsigned value. MessageType receives
several methods that returns string representations of the names and descriptions.

Also, one of the methods is New. It returns a new Message object based on the
MessageType. For example:

	m, err := CONNECT.New()
	msg := m.(*ConnectMessage)

This would return a ConnectMessage struct, but mapped to the Message interface. You can
then type assert it back to a *ConnectMessage.

Another way to create a new ConnectMessage is to call:

	msg := NewConnectMessage()

Every message type has a New function that returns a new message. The list of available
message types are defined as constants below.

As you may have noticed, the second important item is the Message interface. It defines
several methods that are common to all messages, including Name, Len and Type.
Most importantly, it also defines the Encode and Decode methods.

	Encode(dst []byte) (int, error)
	Decode(src []byte) (int, error)

Encode encodes the message into the destination byte slice. The return value is the number
of bytes encoded, so the caller knows how many bytes there will be. If Encode returns an
error, then the encoded bytes should be considered invalid. The destination slice must
have enough capacity for the message.

Decode decodes the byte slice and populates the Message. The first return value is the
number of bytes read. The second is error if Decode encounters any problems.

With these in mind, we can now do:

	// Create a new CONNECT message
	msg := NewConnectMessage()

	// Set the appropriate parameters
	msg.SetWillQos(1)
	msg.SetVersion(4)
	msg.SetCleanSession(true)
	msg.SetClientId([]byte("gomqtt"))
	msg.SetKeepAlive(10)
	msg.SetWillTopic([]byte("will"))
	msg.SetWillMessage([]byte("send me home"))
	msg.SetUsername([]byte("gomqtt"))
	msg.SetPassword([]byte("verysecret"))

	// Allocate a buffer
	buf := make([]byte, msg.Len())

	// Encode the message
	n, err := msg.Encode(buf)
	if err == nil {
		return err
	}

	// Write n bytes into the connection
	m, err := io.CopyN(conn, buf, int64(n))
	if err != nil {
		return err
	}

	fmt.Printf("Sent %d bytes of %s message", m, msg.Name())

To receive a CONNECT message from a connection, we can do:

	// Create a new CONNECT message
	msg := NewConnectMessage()

	// Decode the message
	n, err := msg.Decode(buf)

If you don't know what type of message is coming down the pipe, you can do something like this:

	// Create a buffered IO reader for the connection
	br := bufio.NewReader(conn)

	// Peek at the first byte, which contains the message type
	b, err := br.Peek(1)
	if err != nil {
		return err
	}

	// Extract the type from the first byte
	t := MessageType(b[0] >> 4)

	// Create a new message
	msg, err := t.New()
	if err != nil {
		return err
	}

	// Decode it from the bufio.Reader
	n, err := msg.Decode(br)
	if err != nil {
		return err
	}
*/
package message

import (
	"bytes"
	"regexp"
)

const (
	// QoS 0: At most once delivery
	// The message is delivered according to the capabilities of the underlying network.
	// No response is sent by the receiver and no retry is performed by the sender. The
	// message arrives at the receiver either once or not at all.
	QosAtMostOnce byte = iota

	// QoS 1: At least once delivery
	// This quality of service ensures that the message arrives at the receiver at least once.
	// A QoS 1 PUBLISH Packet has a Packet Identifier in its variable header and is acknowledged
	// by a PUBACK Packet. Section 2.3.1 provides more information about Packet Identifiers.
	QosAtLeastOnce

	// QoS 2: Exactly once delivery
	// This is the highest quality of service, for use when neither loss nor duplication of
	// messages are acceptable. There is an increased overhead associated with this quality of
	// service.
	QosExactlyOnce

	// QosFailure is a return value for a subscription if there's a problem while subscribing
	// to a specific topic.
	QosFailure = 0x80
)

// Message is an interface defined for all MQTT message types.
type Message interface {
	// Name returns a string representation of the message type. Examples include
	// "PUBLISH", "SUBSCRIBE", and others. This is statically defined for each of
	// the message types and cannot be changed.
	Name() string

	// Decode reads the bytes in the byte slice from the argument. It returns the
	// total number of bytes decoded, and whether there have been any errors during
	// the process. The byte slice MUST NOT be modified during the duration of this
	// message being available since the byte slice never gets copied.
	Decode([]byte) (int, error)

	// Encode writes the message bytes into the byte array from the argument. It
	// returns the number of bytes encoded and whether there's any errors along
	// the way. If there's any errors, then the byte slice and count should be
	// considered invalid.
	Encode([]byte) (int, error)

	// Len returns the byte length of the message.
	Len() int
}

/*
A basic fuzzing test that works with https://github.com/dvyukov/go-fuzz:

	$ go-fuzz-build github.com/gomqtt/message
	$ go-fuzz -bin=./message-fuzz.zip -workdir=./fuzz
*/
func Fuzz(data []byte) int {
	// check for zero length data
	if len(data) == 0 {
		return 1
	}

	// Detect message.
	l, mt := DetectMessage(data)

	// Check length
	if l == 0 {
		// for testing purposes we will not cancel
		// on incomplete buffers
	}

	// Create a new message
	msg, err := mt.New()
	if err != nil {
		return 0
	}

	// Decode it from the buffer.
	_, err = msg.Decode(data)
	if err != nil {
		return 0
	}

	// Prepare buffer with 255 as zero value
	buf := make([]byte, msg.Len())
	for i := 0; i < len(buf); i++ {
		buf[i] = 255
	}

	// Try encode the message again.
	_, err = msg.Encode(buf)
	if err != nil {
		// ignore protocol level errors
	}

	// Finally try to decode again.
	_, err = msg.Decode(buf)
	if err != nil {
		return 0
	}

	// Everything was ok!
	return 1
}
