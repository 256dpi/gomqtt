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

// ConnackCode is the type representing the return code in the CONNACK message.
type ConnackCode byte

const (
	// The server has accepted the connection.
	ConnectionAccepted ConnackCode = iota

	// The Server does not support the level of the MQTT protocol requested by the Client.
	ErrInvalidProtocolVersion

	// The Client identifier is correct UTF-8 but not allowed by the server.
	ErrIdentifierRejected

	// The Network Connection has been made but the MQTT service is unavailable.
	ErrServerUnavailable

	// The data in the user name or password is malformed.
	ErrBadUsernameOrPassword

	// The Client is not authorized to connect.
	ErrNotAuthorized
)

// Valid checks to see if the ConnackCode is valid.
func (this ConnackCode) Valid() bool {
	return this <= 5
}

// Error returns the corresponding error string for the ConnackCode.
func (this ConnackCode) Error() string {
	switch this {
	case ConnectionAccepted:
		return "Connection accepted"
	case ErrInvalidProtocolVersion:
		return "Connection Refused, unacceptable protocol version"
	case ErrIdentifierRejected:
		return "Connection Refused, identifier rejected"
	case ErrServerUnavailable:
		return "Connection Refused, Server unavailable"
	case ErrBadUsernameOrPassword:
		return "Connection Refused, bad user name or password"
	case ErrNotAuthorized:
		return "Connection Refused, not authorized"
	}

	return "Unknown error"
}
