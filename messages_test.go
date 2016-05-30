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

package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageString(t *testing.T) {
	msg := &Message{
		Topic:   "w",
		Payload: []byte("m"),
		QOS:     QOSAtLeastOnce,
	}

	assert.Equal(t, "<Message Topic=\"w\" QOS=1 Retain=false Payload=[109]>", msg.String())
}

func TestMessageCopy(t *testing.T) {
	msg1 := &Message{
		Topic:   "w",
		Payload: []byte("m"),
		QOS:     QOSAtLeastOnce,
	}

	msg2 := msg1.Copy()
	assert.Equal(t, msg1.String(), msg2.String())

	msg1.Retain = true
	assert.False(t, msg2.Retain)
}
