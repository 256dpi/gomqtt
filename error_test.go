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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorStrings(t *testing.T) {
	err := &Error{DialError, errors.New("foo")}
	assert.Equal(t, "dial error: foo", err.Error())

	err = &Error{LaunchError, errors.New("foo")}
	assert.Equal(t, "launch error: foo", err.Error())

	err = &Error{EncodeError, errors.New("foo")}
	assert.Equal(t, "encode error: foo", err.Error())

	err = &Error{DecodeError, errors.New("foo")}
	assert.Equal(t, "decode error: foo", err.Error())

	err = &Error{DetectionError, errors.New("foo")}
	assert.Equal(t, "detection error: foo", err.Error())

	err = &Error{NetworkError, errors.New("foo")}
	assert.Equal(t, "network error: foo", err.Error())

	err = &Error{0, errors.New("foo")}
	assert.Equal(t, "unknown error: foo", err.Error())
}
