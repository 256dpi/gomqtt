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

package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Not enough bytes
func TestMessageHeaderDecode(t *testing.T) {
	buf := []byte{0x6f, 193, 2}
	header := &header{}

	_, _, _, err := header.decode(buf)
	require.Error(t, err)
}

// Remaining length too big
func TestMessageHeaderDecode2(t *testing.T) {
	buf := []byte{0x62, 0xff, 0xff, 0xff, 0xff}
	header := &header{}

	_, _, _, err := header.decode(buf)
	require.Error(t, err)
}

func TestMessageHeaderDecode3(t *testing.T) {
	buf := []byte{0x62, 0xff}
	header := &header{}

	_, _, _, err := header.decode(buf)
	require.Error(t, err)
}

func TestMessageHeaderDecode4(t *testing.T) {
	buf := []byte{0x62, 0xff, 0xff, 0xff, 0x7f}
	header := &header{messageType: 6}

	n, _, _, err := header.decode(buf)

	require.Error(t, err)
	require.Equal(t, 5, n)
}

func TestMessageHeaderDecode5(t *testing.T) {
	buf := []byte{0x62, 0xff, 0x7f}
	header := &header{messageType: 6}

	n, _, _, err := header.decode(buf)
	require.Error(t, err)
	require.Equal(t, 3, n)
}

func TestMessageHeaderEncode1(t *testing.T) {
	header := &header{}
	headerBytes := []byte{0x62, 193, 2}

	header.messageType = PUBREL

	buf := make([]byte, 3)
	n, err := header.encode(buf, 0, 321)

	require.NoError(t, err)
	require.Equal(t, 3, n)
	require.Equal(t, headerBytes, buf)
}

func TestMessageHeaderEncode2(t *testing.T) {
	header := &header{}

	header.messageType = PUBREL

	buf := make([]byte, 5)
	_, err := header.encode(buf, 0, 268435456)

	require.Error(t, err)
}

func TestMessageHeaderEncode3(t *testing.T) {
	header := &header{}
	headerBytes := []byte{0x62, 0xff, 0xff, 0xff, 0x7f}

	header.messageType = PUBREL

	buf := make([]byte, 5)
	n, err := header.encode(buf, 0, maxRemainingLength)

	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, headerBytes, buf)
}

func TestMessageHeaderEncode4(t *testing.T) {
	header := &header{messageType: RESERVED2}

	buf := make([]byte, 5)
	_, err := header.encode(buf, 0, 0)
	require.Error(t, err)
}
