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

import "fmt"

type ErrorCode int

const(
	_ ErrorCode = iota
	ExpectedClose
	DialError
	EncodeError
	DecodeError
	DetectionError
	ConnectionError
	ReadLimitExceeded
)

type Error interface {
	error

	// Code will return the corresponding error code.
	Code() ErrorCode

	// Err will return the original error.
	Err() error
}

type transportError struct {
	code ErrorCode
	err error
}

func newTransportError(code ErrorCode, err error) *transportError {
	return &transportError{
		code: code,
		err: err,
	}
}

func (err *transportError) Error() string {
	switch err.code {
	case ExpectedClose:
		return fmt.Sprintf("expected close: %s", err.err.Error())
	case DialError:
		return fmt.Sprintf("dial error: %s", err.err.Error())
	case EncodeError:
		return fmt.Sprintf("encode error: %s", err.err.Error())
	case DecodeError:
		return fmt.Sprintf("decode error: %s", err.err.Error())
	case DetectionError:
		return fmt.Sprintf("detection error: %s", err.err.Error())
	case ConnectionError:
		return fmt.Sprintf("connection error: %s", err.err.Error())
	case ReadLimitExceeded:
		return fmt.Sprintf("read limit exceeded: %s", err.err.Error())
	}

	return fmt.Sprintf("unknown error: %s", err.err.Error())
}

func (err *transportError) Code() ErrorCode {
	return err.code
}

func (err *transportError) Err() error {
	return err.err
}
