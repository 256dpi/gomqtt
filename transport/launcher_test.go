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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalLaunch(t *testing.T) {
	server, err := Launch("tcp://localhost:0")
	require.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func TestLauncherBadURL(t *testing.T) {
	conn, err := Launch("foo")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestLauncherUnsupportedProtocol(t *testing.T) {
	conn, err := Launch("foo://localhost")
	assert.Nil(t, conn)
	assert.Equal(t, ErrUnsupportedProtocol, err)
}
