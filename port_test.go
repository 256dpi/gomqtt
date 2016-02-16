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

package tools

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPort(t *testing.T) {
	port := NewPort()
	assert.Equal(t, fmt.Sprintf("tcp://localhost:%s/", port.Port()), port.URL())
	assert.Equal(t, fmt.Sprintf("ws://localhost:%s/", port.Port()), port.URL("ws"))
}
