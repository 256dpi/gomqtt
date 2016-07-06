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

package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/gomqtt/spec"
)

func TestBrokerWithMemoryBackend(t *testing.T) {
	defer leaktest.Check(t)()

	backend := NewMemoryBackend()
	backend.Logins = map[string]string{
		"allow": "allow",
	}

	port, quit, done := Run(t, NewEngineWithBackend(backend), "tcp")

	config := spec.AllFeatures()
	config.URL = fmt.Sprintf("tcp://allow:allow@localhost:%s", port.Port())
	config.DenyURL = fmt.Sprintf("tcp://deny:deny@localhost:%s", port.Port())
	config.NoMessageWait = 50 * time.Millisecond
	config.MessageRetainWait = 50 * time.Millisecond

	spec.Run(t, config)

	close(quit)

	<-done
}
