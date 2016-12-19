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

package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testStrings = []string{
		"this is a test",
		"hope it succeeds",
		"but just in case",
		"send me your millions",
		"",
	}

	testBytes = []byte{
		0x0, 0xe, 't', 'h', 'i', 's', ' ', 'i', 's', ' ', 'a', ' ', 't', 'e', 's', 't',
		0x0, 0x10, 'h', 'o', 'p', 'e', ' ', 'i', 't', ' ', 's', 'u', 'c', 'c', 'e', 'e', 'd', 's',
		0x0, 0x10, 'b', 'u', 't', ' ', 'j', 'u', 's', 't', ' ', 'i', 'n', ' ', 'c', 'a', 's', 'e',
		0x0, 0x15, 's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'y', 'o', 'u', 'r', ' ', 'm', 'i', 'l', 'l', 'i', 'o', 'n', 's',
		0x0, 0x0,
	}
)

func TestReadLPBytes(t *testing.T) {
	total := 0

	for _, str := range testStrings {
		b, n, err := readLPBytes(testBytes[total:], true, CONNECT)

		assert.NoError(t, err)
		assert.Equal(t, []byte(str), b)
		assert.Equal(t, len(str)+2, n)

		total += n
	}
}

func TestReadLPBytesErrors(t *testing.T) {
	_, _, err := readLPBytes([]byte{}, true, CONNECT)
	assert.Error(t, err)

	_, _, err = readLPBytes([]byte{0xff, 0xff, 0xff, 0xff}, true, CONNECT)
	assert.Error(t, err)
}

func TestReadLPStringErrors(t *testing.T) {
	_, _, err := readLPString([]byte{}, CONNECT)
	assert.Error(t, err)

	_, _, err = readLPString([]byte{0xff, 0xff, 0xff, 0xff}, CONNECT)
	assert.Error(t, err)
}

func TestWriteLPBytes(t *testing.T) {
	total := 0
	buf := make([]byte, 127)

	for _, str := range testStrings {
		n, err := writeLPBytes(buf[total:], []byte(str), CONNECT)

		assert.NoError(t, err)
		assert.Equal(t, 2+len(str), n)

		total += n
	}

	assert.Equal(t, testBytes, buf[:total])
}

func TestWriteLPBytesErrors(t *testing.T) {
	_, err := writeLPBytes([]byte{}, make([]byte, 65536), CONNECT)
	assert.Error(t, err)

	_, err = writeLPBytes([]byte{}, make([]byte, 10), CONNECT)
	assert.Error(t, err)
}
