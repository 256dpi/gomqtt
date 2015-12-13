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
	"fmt"
	"strings"
	"sync"
)

type node struct {
	children map[string]*node
	values   []interface{}
}

func newNode() *node {
	return &node{
		children: make(map[string]*node),
		values:   make([]interface{}, 0, 1),
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

// Tree implements a thread-safe tree for adding, removing and matching topics.
type Tree struct {
	// The separator character. Default: "/"
	Separator    string

	// The single level wildcard character. Default: "+"
	WildcardOne  string

	// The multi level wildcard character. Default "#"
	WildcardSome string

	root  *node
	mutex sync.RWMutex
}

// NewTree returns a new Tree.
func NewTree() *Tree {
	return &Tree{
		Separator:    "/",
		WildcardOne:  "+",
		WildcardSome: "#",

		root: newNode(),
	}
}

// Add registers the value for the supplied filter. This function will
// automatically grow the tree and is thread-safe.
func (t *Tree) Add(filter string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

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

	t.add(value, i+1, segments, subtree)
}

// Remove unregisters the value for the supplied filter. This function will
// automatically shrink the tree and is thread-safe.
func (t *Tree) Remove(filter string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.remove(value, 0, strings.Split(filter, t.Separator), t.root)
}

// Empty will unregister all values for the supplied filter. This function will
// automatically shrink the tree and is thread-safe.
func (t *Tree) Empty(filter string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

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

	if t.remove(value, i+1, segments, subtree) {
		delete(tree.children, segment)
	}

	return len(tree.values) == 0 && len(tree.children) == 0
}

// Clear will unregister the supplied value from all filters. This function will
// automatically shrink the tree and is thread-safe.
func (t *Tree) Clear(value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

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

// Match will return a set of values that have filters that match the supplied
// topic. The result set will be cleared from duplicate values. The function is
// thread-safe.
func (t *Tree) Match(topic string) []interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	values := t.match(make([]interface{}, 0), 0, strings.Split(topic, t.Separator), t.root)

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
	// when finished add all values to the result set
	if i == len(segments) {
		return append(result, tree.values...)
	}

	// add all values to the result set that match multiple levels
	if subtree, ok := tree.children[t.WildcardSome]; ok {
		result = append(result, subtree.values...)
	}

	// advance children that match a single level
	if subtree, ok := tree.children[t.WildcardOne]; ok {
		result = t.match(result, i+1, segments, subtree)
	}

	segment := segments[i]

	// match segments and get children
	if segment != t.WildcardOne && segment != t.WildcardSome {
		if subtree, ok := tree.children[segment]; ok {
			result = t.match(result, i+1, segments, subtree)
		}
	}

	return result
}

// Reset will completely clear the tree. The function is thread-safe.
func (t *Tree) Reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.root = newNode()
}

// String will return a string representation of the tree. The function is
// thread-safe.
func (t *Tree) String() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return fmt.Sprintf("topic.Tree:%s", t.root.string(0))
}
