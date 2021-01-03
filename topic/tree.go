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
	}
}

func (n *node) removeValue(value interface{}) {
	for i, v := range n.values {
		if v == value {
			// remove without preserving order
			n.values[i] = n.values[len(n.values)-1]
			n.values[len(n.values)-1] = nil
			n.values = n.values[:len(n.values)-1]
			break
		}
	}
}

func (n *node) clearValues() {
	n.values = []interface{}{}
}

func (n *node) string(level int) string {
	// print node length unless on root level
	str := ""
	if level != 0 {
		str = fmt.Sprintf("%d", len(n.values))
	}

	// ident and append children
	for key, node := range n.children {
		str += fmt.Sprintf("\n| %s'%s' => %s", strings.Repeat(" ", level*2), key, node.string(level+1))
	}

	return str
}

// A Tree implements a thread-safe topic tree.
type Tree struct {
	root  *node
	mutex sync.RWMutex
}

// NewTree returns a new Tree.
func NewTree() *Tree {
	return &Tree{
		root: newNode(),
	}
}

// Add registers the value for the supplied topic. This function will
// automatically grow the tree. If value already exists for the given topic it
// will not be added again.
func (t *Tree) Add(topic string, value interface{}) {
	// acquire mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// add value
	t.add(value, topic, t.root)
}

func (t *Tree) add(value interface{}, topic string, node *node) {
	// add value to leaf node
	if topic == topicEnd {
		// add value if not already added
		if !contains(node.values, value) {
			node.values = append(node.values, value)
		}

		return
	}

	// get segment
	segment := topicSegment(topic)

	// get child
	child, ok := node.children[segment]
	if !ok {
		child = newNode()
		node.children[segment] = child
	}

	// descend
	t.add(value, topicShorten(topic), child)
}

// Set sets the supplied value as the only value for the supplied topic. This
// function will automatically grow the tree.
func (t *Tree) Set(topic string, value interface{}) {
	// acquire mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// set value
	t.set(value, topic, t.root)
}

func (t *Tree) set(value interface{}, topic string, node *node) {
	// set value on leaf node
	if topic == topicEnd {
		node.values = []interface{}{value}
		return
	}

	// get segment
	segment := topicSegment(topic)

	// get child
	child, ok := node.children[segment]
	if !ok {
		child = newNode()
		node.children[segment] = child
	}

	// descend
	t.set(value, topicShorten(topic), child)
}

// Get gets the values from the topic that exactly matches the supplied topics.
func (t *Tree) Get(topic string) []interface{} {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// get values
	return t.get(topic, t.root)
}

func (t *Tree) get(topic string, node *node) []interface{} {
	// return values from leaf node
	if topic == topicEnd {
		return node.values
	}

	// get segment
	segment := topicSegment(topic)

	// get child
	child, ok := node.children[segment]
	if !ok {
		return nil
	}

	// descend
	return t.get(topicShorten(topic), child)
}

// Remove un-registers the value from the supplied topic. This function will
// automatically shrink the tree.
func (t *Tree) Remove(topic string, value interface{}) {
	// acquire mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// remove value
	t.remove(value, topic, t.root)
}

// Empty will unregister all values from the supplied topic. This function will
// automatically shrink the tree.
func (t *Tree) Empty(topic string) {
	// acquire mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// empty values
	t.remove(nil, topic, t.root)
}

func (t *Tree) remove(value interface{}, topic string, node *node) bool {
	// clear or remove value from leaf node
	if topic == topicEnd {
		if value == nil {
			node.clearValues()
		} else {
			node.removeValue(value)
		}

		return len(node.values) == 0 && len(node.children) == 0
	}

	// get segment
	segment := topicSegment(topic)

	// get child
	child, ok := node.children[segment]
	if !ok {
		return false
	}

	// descend and remove node if empty
	if t.remove(value, topicShorten(topic), child) {
		delete(node.children, segment)
	}

	return len(node.values) == 0 && len(node.children) == 0
}

// Clear will unregister the supplied value from all topics. This function will
// automatically shrink the tree.
func (t *Tree) Clear(value interface{}) {
	// acquire mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// clear value
	t.clear(value, t.root)
}

func (t *Tree) clear(value interface{}, node *node) bool {
	// remove value
	node.removeValue(value)

	// remove value from all children and remove empty nodes
	for segment, child := range node.children {
		if t.clear(value, child) {
			delete(node.children, segment)
		}
	}

	return len(node.values) == 0 && len(node.children) == 0
}

// Match will return a set of values from topics that match the supplied topic.
// The result set will be cleared from duplicate values.
//
// Note: In contrast to Search, Match does not respect wildcards in the query but
// in the stored tree.
func (t *Tree) Match(topic string) []interface{} {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// match values
	var list []interface{}
	t.match(topic, t.root, func(values []interface{}) bool {
		list = append(list, values...)
		return true
	})

	return unique(list)
}

// MatchFirst behaves similar to Match but only returns the first found value.
func (t *Tree) MatchFirst(topic string) interface{} {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// match value
	var value interface{}
	t.match(topic, t.root, func(values []interface{}) bool {
		value = values[0]
		return false
	})

	return value
}

func (t *Tree) match(topic string, node *node, fn func([]interface{}) bool) {
	// add all values to the result set that match multiple levels
	if child, ok := node.children["#"]; ok && len(child.values) > 0 {
		if !fn(child.values) {
			return
		}
	}

	// when finished add all values to the result set
	if topic == topicEnd {
		if len(node.values) > 0 {
			fn(node.values)
		}

		return
	}

	// advance children that match a single level
	if child, ok := node.children["+"]; ok {
		t.match(topicShorten(topic), child, fn)
	}

	// get segment
	segment := topicSegment(topic)

	// match segments and get children
	if segment != "+" && segment != "#" {
		if child, ok := node.children[segment]; ok {
			t.match(topicShorten(topic), child, fn)
		}
	}
}

// Search will return a set of values from topics that match the supplied topic.
// The result set will be cleared from duplicate values.
//
// Note: In contrast to Match, Search respects wildcards in the query but not in
// the stored tree.
func (t *Tree) Search(topic string) []interface{} {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// match values
	var list []interface{}
	t.search(topic, t.root, func(values []interface{}) bool {
		list = append(list, values...)
		return true
	})

	return unique(list)
}

// SearchFirst behaves similar to Search but only returns the first found value.
func (t *Tree) SearchFirst(topic string) interface{} {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// match value
	var value interface{}
	t.search(topic, t.root, func(values []interface{}) bool {
		value = values[0]
		return false
	})

	return value
}

func (t *Tree) search(topic string, node *node, fn func([]interface{}) bool) {
	// when finished add all values to the result set
	if topic == topicEnd {
		if len(node.values) > 0 {
			fn(node.values)
		}

		return
	}

	// get segment
	segment := topicSegment(topic)

	// add all current and further values
	if segment == "#" {
		if len(node.values) > 0 {
			if !fn(node.values) {
				return
			}
		}

		for _, child := range node.children {
			t.search(topic, child, fn)
		}
	}

	// add all current values and continue
	if segment == "+" {
		if len(node.values) > 0 {
			if !fn(node.values) {
				return
			}
		}

		for _, child := range node.children {
			t.search(topicShorten(topic), child, fn)
		}
	}

	// match segments and get children
	if segment != "+" && segment != "#" {
		if child, ok := node.children[segment]; ok {
			t.search(topicShorten(topic), child, fn)
		}
	}
}

// Count will count all stored values in the tree. It will not filter out
// duplicate values and thus might return a different result to `len(All())`.
func (t *Tree) Count() int {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.count(t.root)
}

func (t *Tree) count(node *node) int {
	// prepare total
	total := 0

	// add children to results
	for _, child := range node.children {
		total += t.count(child)
	}

	// add values to result
	return total + len(node.values)
}

// All will return all stored values in the tree.
func (t *Tree) All() []interface{} {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return unique(t.all([]interface{}{}, t.root))
}

func (t *Tree) all(result []interface{}, node *node) []interface{} {
	// add children to results
	for _, child := range node.children {
		result = t.all(result, child)
	}

	// add current node to results
	return append(result, node.values...)
}

// Reset will completely clear the tree.
func (t *Tree) Reset() {
	// acquire mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// set new root
	t.root = newNode()
}

// String will return a string representation of the tree structure. The number
// following the nodes show the number of stored values at that level.
func (t *Tree) String() string {
	// acquire mutex
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return fmt.Sprintf("topic.Tree:%s", t.root.string(0))
}

func unique(values []interface{}) []interface{} {
	// collect unique values
	result := values[:0]
	for _, v := range values {
		if !contains(result, v) {
			result = append(result, v)
		}
	}

	return result
}

func contains(list []interface{}, value interface{}) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

var topicEnd = "\x00"

func topicShorten(topic string) string {
	i := strings.IndexRune(topic, '/')
	if i >= 0 {
		return topic[i+1:]
	}

	return topicEnd
}

func topicSegment(topic string) string {
	i := strings.IndexRune(topic, '/')
	if i >= 0 {
		return topic[:i]
	}

	return topic
}
