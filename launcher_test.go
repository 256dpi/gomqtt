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

	"github.com/stretchr/testify/require"
)

func TestGlobalLaunch(t *testing.T) {
	tp := newTestPort()

	server, err := Launch(tp.url("tcp"))
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)
}

func TestLauncherBadURL(t *testing.T) {
	conn, err := Launch("foo")
	require.Nil(t, conn)
	require.Equal(t, LaunchError, toError(err).Code())
}

func TestLauncherUnsupportedProtocol(t *testing.T) {
	conn, err := Launch("foo://localhost")
	require.Nil(t, conn)
	require.Equal(t, LaunchError, toError(err).Code())
	require.Equal(t, ErrUnsupportedProtocol, toError(err).Err())
}
