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

// Package packet implements functionality for encoding and decoding MQTT 3.1.1
// (http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/) packets.
package packet

import "encoding/binary"

const (
	// QOSAtMostOnce defines that the message is delivered at most once, or it
	// may not be delivered at all.
	QOSAtMostOnce byte = iota

	// QOSAtLeastOnce defines that the message is always delivered at least once.
	QOSAtLeastOnce

	// QOSExactlyOnce defines that the message is always delivered exactly once.
	QOSExactlyOnce

	// QOSFailure indicates that there has been an error while subscribing
	// to a specific topic.
	QOSFailure = 0x80
)

// Packet is an abstract interface for all MQTT packet types.
type Packet interface {
	// Type returns the packets type.
	Type() Type

	// Len returns the byte length of the encoded packet.
	Len() int

	// Decode reads from the byte slice argument. It returns the total number of
	// bytes decoded, and whether there have been any errors during the process.
	// The byte slice must not be modified during the duration of this packet
	// being available since the byte slice never gets copied.
	Decode([]byte) (int, error)

	// Encode writes the packet bytes into the byte slice from the argument. It
	// returns the number of bytes encoded and whether there's any errors along
	// the way. If there is an error, the byte slice should be considered invalid.
	Encode([]byte) (int, error)

	// String returns a string representation of the packet.
	String() string
}

// DetectPacket tries to detect the next packet in a buffer. It returns a length
// greater than zero if the packet has been detected as well as its Type.
func DetectPacket(src []byte) (int, Type) {
	// check for minimum size
	if len(src) < 2 {
		return 0, 0
	}

	// get type
	t := Type(src[0] >> 4)

	// get remaining length
	_rl, n := binary.Uvarint(src[1:])
	rl := int(_rl)

	if n <= 0 {
		return 0, 0
	}

	return 1 + n + rl, t
}

// Fuzz is a basic fuzzing test that works with https://github.com/dvyukov/go-fuzz:
//
//		$ go-fuzz-build github.com/gomqtt/packet
//		$ go-fuzz -bin=./packet-fuzz.zip -workdir=./fuzz
func Fuzz(data []byte) int {
	// check for zero length data
	if len(data) == 0 {
		return 1
	}

	// Detect packet.
	_, mt := DetectPacket(data)

	// For testing purposes we will not cancel
	// on incomplete buffers

	// Create a new packet
	pkt, err := mt.New()
	if err != nil {
		return 0
	}

	// Decode it from the buffer.
	_, err = pkt.Decode(data)
	if err != nil {
		return 0
	}

	// Everything was ok!
	return 1
}
