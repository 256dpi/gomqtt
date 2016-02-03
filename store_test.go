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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gomqtt/packet"
)

func TestMemoryStore1(t *testing.T) {
	store := NewMemoryStore()

	err := store.Open()
	assert.NoError(t, err)

	pkt1 := packet.NewPublishPacket()
	pkt1.PacketID = 1

	err = store.Put(pkt1)
	assert.NoError(t, err)

	pkt2, err := store.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, pkt1, pkt2)

	err = store.Del(1)
	assert.NoError(t, err)

	pkt3, err := store.Get(1)
	assert.NoError(t, err)
	assert.Nil(t, pkt3)

	err = store.Close()
	assert.NoError(t, err)
}

func TestMemoryStore2(t *testing.T) {
	store := NewMemoryStore()

	err := store.Open()
	assert.NoError(t, err)

	pkt1 := packet.NewPublishPacket()
	pkt1.PacketID = 1

	err = store.Put(pkt1)
	assert.NoError(t, err)

	pkts, err := store.All()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pkts))

	err = store.Reset()
	assert.NoError(t, err)

	pkts, err = store.All()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pkts))

	err = store.Close()
	assert.NoError(t, err)
}
