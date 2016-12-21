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
)

func TestTCPServer(t *testing.T) {
	abstractServerTest(t, "tcp")
}

func TestTLSServer(t *testing.T) {
	abstractServerTest(t, "tls")
}

func TestTCPServerLaunchError(t *testing.T) {
	abstractServerLaunchErrorTest(t, "tcp")
}

func TestTLSServerLaunchError(t *testing.T) {
	abstractServerLaunchErrorTest(t, "tls")
}

func TestNetServerAcceptAfterClose(t *testing.T) {
	abstractServerAcceptAfterCloseTest(t, "tcp")
}

func TestNetServerCloseAfterClose(t *testing.T) {
	abstractServerCloseAfterCloseTest(t, "tcp")
}

func TestNetServerAddr(t *testing.T) {
	abstractServerAddrTest(t, "tcp")
}
