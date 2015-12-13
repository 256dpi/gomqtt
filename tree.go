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
	"strings"
	"fmt"
)

type node struct {
	children map[string]*node
	values []interface{}
}

func newNode() *node {
	return &node{
		children: make(map[string]*node),
		values: make([]interface{}, 0, 1),
	}
}

func (n *node) removeValue(value interface{}) {
	for i, v := range n.values {
		if v == value {
			// remove without preserving order
			n.values[i] = n.values[len(n.values)-1]
			n.values = n.values[:len(n.values)-1]
			break
		}
	}
}

func (n *node) clearValues() {
	n.values = make([]interface{}, 0, 1)
}

func (n *node) string(i int) string {
	str := ""

	if i != 0 {
		str = fmt.Sprintf("%d", len(n.values))
	}

	for key, node := range n.children {
		str += fmt.Sprintf("\n| %s'%s' => %s", strings.Repeat(" ", i*2), key, node.string(i+1))
	}

	return str
}

type Tree struct {
	Separator string
	WildcardOne string
	WildcardSome string

	root *node
}

func NewTree() *Tree {
	return &Tree{
		Separator: "/",
		WildcardOne: "+",
		WildcardSome: "#",

		root: newNode(),
	}
}

func (t *Tree) Add(filter string, value interface{}) {
	t.add(value, 0, strings.Split(filter, t.Separator), t.root)
}

func (t *Tree) add(value interface{}, i int, segments []string, tree *node) {
	// add value to leaf
	if i == len(segments) {
		tree.values = append(tree.values, value)
		return
	}

	segment := segments[i]
	subtree, ok := tree.children[segment]

	// create missing node
	if !ok {
		subtree = newNode()
		tree.children[segment] = subtree
	}

	t.add(value, i + 1, segments, subtree)
}

func (t *Tree) Remove(filter string, value interface{}) {
	t.remove(value, 0, strings.Split(filter, t.Separator), t.root)
}

func (t *Tree) Empty(filter string) {
	t.remove(nil, 0, strings.Split(filter, t.Separator), t.root)
}

func (t *Tree) remove(value interface{}, i int, segments []string, tree *node) bool {
	// clear or remove value from leaf node
	if i == len(segments) {
		if value == nil {
			tree.clearValues()
		} else {
			tree.removeValue(value)
		}

		return len(tree.values) == 0 && len(tree.children) == 0
	}

	segment := segments[i]
	subtree, ok := tree.children[segment]

	// node not found
	if !ok {
		return false
	}

	if t.remove(value, i + 1, segments, subtree) {
		delete(tree.children, segment)
	}

	return len(tree.values) == 0 && len(tree.children) == 0
}

func (t *Tree) Clear(value interface{}) {
	t.clear(value, t.root)
}

func (t *Tree) clear(value interface{}, tree *node) bool {
	tree.removeValue(value)

	// remove value from all nodes
	for segment, subtree := range tree.children {
		if t.clear(value, subtree) {
			delete(tree.children, segment)
		}
	}

	return len(tree.values) == 0 && len(tree.children) == 0
}

func (t *Tree) Match(filter string) ([]interface{}) {
	values := t.match(make([]interface{}, 0), 0, strings.Split(filter, t.Separator), t.root)

	// remove duplicates
	result := make([]interface{}, 0, len(values))
	seen := make(map[interface{}]bool, len(values))
	for _, v := range values {
		if _, ok := seen[v]; !ok {
			result = append(result, v)
			seen[v] = true
		}
	}
	return result
}

func (t *Tree) match(result []interface{}, i int, segments []string, tree *node) []interface{} {
	// add all values for the some wildcard to the result set
	if subtree, ok := tree.children[t.WildcardSome]; ok {
		result = append(result, subtree.values...)
	}

	// advance children of the one wildcard
	if subtree, ok := tree.children[t.WildcardOne]; ok {
		result = t.match(result, i + 1, segments, subtree)
	}

	// when finished add all values to the result set
	if i == len(segments) {
		return append(result, tree.values...)
	}

	segment := segments[i]

	// match segments and get children
	if segment != t.WildcardOne && segment != t.WildcardSome {
		if subtree, ok := tree.children[segment]; ok {
			result = t.match(result, i + 1, segments, subtree)
		}
	}

	return result
}

func (t* Tree) Reset() {
	t.root = newNode()
}

func (t* Tree) String() string {
	return fmt.Sprintf("topic.Tree:%s\n", t.root.string(0))
}
