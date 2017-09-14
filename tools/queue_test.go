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
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueue(t *testing.T) {
	msg := &packet.Message{}

	queue := NewQueue(2)
	assert.Equal(t, 0, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 1, queue.Len())

	msg1 := queue.Pop()
	assert.Equal(t, msg, msg1)
	assert.Equal(t, 0, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 1, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 2, queue.Len())

	queue.Push(msg)
	assert.Equal(t, 2, queue.Len())

	msgs := queue.All()
	assert.Equal(t, 0, queue.Len())
	assert.Equal(t, []*packet.Message{msg, msg}, msgs)

	msg2 := queue.Pop()
	assert.Nil(t, msg2)
	assert.Equal(t, 0, queue.Len())
}
