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

package transport

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/require"
)

func TestErrorStrings(t *testing.T) {
	err1 := newTransportError(ExpectedClose, fmt.Errorf("foo"))
	require.Equal(t, "expected close: foo", err1.Error())

	err2 := newTransportError(DialError, fmt.Errorf("foo"))
	require.Equal(t, "dial error: foo", err2.Error())

	err3 := newTransportError(EncodeError, fmt.Errorf("foo"))
	require.Equal(t, "encode error: foo", err3.Error())

	err4 := newTransportError(DecodeError, fmt.Errorf("foo"))
	require.Equal(t, "decode error: foo", err4.Error())

	err5 := newTransportError(DetectionError, fmt.Errorf("foo"))
	require.Equal(t, "detection error: foo", err5.Error())

	err6 := newTransportError(ConnectionError, fmt.Errorf("foo"))
	require.Equal(t, "connection error: foo", err6.Error())

	err7 := newTransportError(ReadLimitExceeded, fmt.Errorf("foo"))
	require.Equal(t, "read limit exceeded: foo", err7.Error())

	err8 := newTransportError(0, fmt.Errorf("foo"))
	require.Equal(t, "unknown error: foo", err8.Error())
}
