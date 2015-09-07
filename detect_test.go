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

func TestDetect1(t *testing.T) {
	buf := []byte{0x10, 0x0}

	l, mt := DetectMessage(buf)

	require.Equal(t, 2, l)
	require.Equal(t, 1, int(mt))
}

// not enough bytes
func TestDetect2(t *testing.T) {
	buf := []byte{0x10, 0xff}

	l, mt := DetectMessage(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(mt))
}

func TestDetect3(t *testing.T) {
	buf := []byte{0x10, 0xff, 0x0}

	l, mt := DetectMessage(buf)

	require.Equal(t, 130, l)
	require.Equal(t, 1, int(mt))
}

// not enough bytes
func TestDetect4(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff}

	l, mt := DetectMessage(buf)

	require.Equal(t, 0, l)
	require.Equal(t, 0, int(mt))
}

func TestDetect5(t *testing.T) {
	buf := []byte{0x10, 0xff, 0xff, 0xff, 0x1}

	l, mt := DetectMessage(buf)

	require.Equal(t, 4194308, l)
	require.Equal(t, 1, int(mt))
}
