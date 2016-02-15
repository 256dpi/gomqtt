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
	"testing"

	"github.com/stretchr/testify/assert"
)

func abstractBackendGetSessionTest(t *testing.T, backend Backend) {
	session1, err := backend.GetSession("foo")
	assert.NoError(t, err)
	assert.NotNil(t, session1)

	session2, err := backend.GetSession("foo")
	assert.NoError(t, err)
	assert.True(t, session1 == session2)

	session3, err := backend.GetSession("bar")
	assert.NoError(t, err)
	assert.False(t, session3 == session1)
	assert.False(t, session3 == session2)
}

// store and look up retained messages
// look up retained messages with a # pattern
// look up retained messages with a + pattern
// remove retained message
// storing twice a retained message should keep only the last

// store and look up subscriptions by client
// remove subscriptions by client
// store and look up subscriptions by topic
// QoS 0 subscriptions, restored but not matched
// clean subscriptions
// store and count subscriptions
