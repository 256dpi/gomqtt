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

package topic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTreeAdd(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)

	require.Equal(t, 1, tree.root.children["foo"].children["bar"].values[0])
}

func TestTreeRemove(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)
	tree.Remove("foo/bar", 1)

	require.Equal(t, 0, len(tree.root.children))
}

func TestTreeRemoveMissing(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)
	tree.Remove("bar/baz", 1)

	require.Equal(t, 1, len(tree.root.children))
}

func TestTreeEmpty(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)
	tree.Add("foo/bar", 2)
	tree.Empty("foo/bar")

	require.Equal(t, 0, len(tree.root.children))
}

func TestTreeClear(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)
	tree.Add("foo/bar/baz", 1)
	tree.Clear(1)

	require.Equal(t, 0, len(tree.root.children))
}

func TestTreeMatchExact(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)

	require.Equal(t, 1, tree.Match("foo/bar")[0])
}

func TestTreeMatchWildcard1(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/+", 1)

	require.Equal(t, 1, tree.Match("foo/bar")[0])
}

func TestTreeMatchWildcard2(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/#", 1)

	require.Equal(t, 1, tree.Match("foo/bar")[0])
}

func TestTreeMatchWildcard3(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/#", 1)

	require.Equal(t, 1, tree.Match("foo/bar/baz")[0])
}

func TestTreeMatchNoDuplicates(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)
	tree.Add("foo/*", 1)
	tree.Add("foo/#", 1)

	require.Equal(t, 1, len(tree.Match("foo/bar")))
}

func TestTreeReset(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)
	tree.Reset()

	require.Equal(t, 0, len(tree.root.children))
}

func TestTreeString(t *testing.T) {
	tree := NewTree()

	tree.Add("foo/bar", 1)

	require.Equal(t, "topic.Tree:\n| 'foo' => 0\n|   'bar' => 1\n", tree.String())
}

func BenchmarkTreeAdd(b *testing.B) {
	tree := NewTree()

	for i := 0; i < b.N; i++ {
		tree.Add("foo/bar", 1)
	}
}

func BenchmarkTreeMatch(b *testing.B) {
	tree := NewTree()
	tree.Add("foo/bar", 1)

	for i := 0; i < b.N; i++ {
		tree.Match("foo/*")
	}
}
