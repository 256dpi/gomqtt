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
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	lpstrings []string = []string{
		"this is a test",
		"hope it succeeds",
		"but just in case",
		"send me your millions",
		"",
	}

	lpstringBytes []byte = []byte{
		0x0, 0xe, 't', 'h', 'i', 's', ' ', 'i', 's', ' ', 'a', ' ', 't', 'e', 's', 't',
		0x0, 0x10, 'h', 'o', 'p', 'e', ' ', 'i', 't', ' ', 's', 'u', 'c', 'c', 'e', 'e', 'd', 's',
		0x0, 0x10, 'b', 'u', 't', ' ', 'j', 'u', 's', 't', ' ', 'i', 'n', ' ', 'c', 'a', 's', 'e',
		0x0, 0x15, 's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'y', 'o', 'u', 'r', ' ', 'm', 'i', 'l', 'l', 'i', 'o', 'n', 's',
		0x0, 0x0,
	}
)

func TestReadLPBytes(t *testing.T) {
	total := 0

	for _, str := range lpstrings {
		b, n, err := readLPBytes(lpstringBytes[total:])

		require.NoError(t, err)
		require.Equal(t, str, string(b))
		require.Equal(t, len(str)+2, n)

		total += n
	}
}

func TestWriteLPBytes(t *testing.T) {
	total := 0
	buf := make([]byte, 127)

	for _, str := range lpstrings {
		n, err := writeLPBytes(buf[total:], []byte(str))

		require.NoError(t, err)
		require.Equal(t, 2+len(str), n)

		total += n
	}

	require.Equal(t, lpstringBytes, buf[:total])
}
