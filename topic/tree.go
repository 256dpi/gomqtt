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
			n.values = n.values[:len(n.values)-1]
			break
		}
	}
}

func (n *node) clearValues() {
	n.values = []interface{}{}
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

// A Tree implements a thread-safe topic tree.
type Tree struct {
	// The separator character. Default: "/"
	Separator string

	// The single level wildcard character. Default: "+"
	WildcardOne string

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

// Add registers the value for the supplied topic. This function will
// automatically grow the tree. If value already exists for the given topic it
// will not be added again.
func (t *Tree) Add(topic string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.add(value, topic, t.root)
}

func (t *Tree) add(value interface{}, topic string, node *node) {
	// add value to leaf
	if topic == topicEnd {
		for _, v := range node.values {
			if v == value {
				return
			}
		}

		node.values = append(node.values, value)
		return
	}

	segment := topicSegment(topic, t.Separator)
	child, ok := node.children[segment]

	// create missing node
	if !ok {
		child = newNode()
		node.children[segment] = child
	}

	t.add(value, topicShorten(topic, t.Separator), child)
}

// Set sets the supplied value as the only value for the supplied topic. This
// function will automatically grow the tree.
func (t *Tree) Set(topic string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.set(value, topic, t.root)
}

func (t *Tree) set(value interface{}, topic string, node *node) {
	// set value on leaf
	if topic == topicEnd {
		node.values = []interface{}{value}
		return
	}

	segment := topicSegment(topic, t.Separator)
	child, ok := node.children[segment]

	// create missing node
	if !ok {
		child = newNode()
		node.children[segment] = child
	}

	t.set(value, topicShorten(topic, t.Separator), child)
}

// Get gets the values from the topic that exactly matches the supplied topics.
func (t *Tree) Get(topic string) []interface{} {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.get(topic, t.root)
}

func (t *Tree) get(topic string, node *node) []interface{} {
	// set value on leaf
	if topic == topicEnd {
		return node.values
	}

	// get next segment
	segment := topicSegment(topic, t.Separator)
	child, ok := node.children[segment]
	if !ok {
		return nil
	}

	return t.get(topicShorten(topic, t.Separator), child)
}

// Remove un-registers the value from the supplied topic. This function will
// automatically shrink the tree.
func (t *Tree) Remove(topic string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.remove(value, topic, t.root)
}

// Empty will unregister all values from the supplied topic. This function will
// automatically shrink the tree.
func (t *Tree) Empty(topic string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

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

	segment := topicSegment(topic, t.Separator)
	child, ok := node.children[segment]

	// node not found
	if !ok {
		return false
	}

	if t.remove(value, topicShorten(topic, t.Separator), child) {
		delete(node.children, segment)
	}

	return len(node.values) == 0 && len(node.children) == 0
}

// Clear will unregister the supplied value from all topics. This function will
// automatically shrink the tree.
func (t *Tree) Clear(value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.clear(value, t.root)
}

func (t *Tree) clear(value interface{}, node *node) bool {
	node.removeValue(value)

	// remove value from all nodes
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
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	values := t.match([]interface{}{}, topic, t.root)

	return t.clean(values)
}

func (t *Tree) match(result []interface{}, topic string, node *node) []interface{} {
	// add all values to the result set that match multiple levels
	if child, ok := node.children[t.WildcardSome]; ok {
		result = append(result, child.values...)
	}

	// when finished add all values to the result set
	if topic == topicEnd {
		return append(result, node.values...)
	}

	// advance children that match a single level
	if child, ok := node.children[t.WildcardOne]; ok {
		result = t.match(result, topicShorten(topic, t.Separator), child)
	}

	// get segment
	segment := topicSegment(topic, t.Separator)

	// match segments and get children
	if segment != t.WildcardOne && segment != t.WildcardSome {
		if child, ok := node.children[segment]; ok {
			result = t.match(result, topicShorten(topic, t.Separator), child)
		}
	}

	return result
}

// MatchFirst will run Match and return the first value or nil.
func (t *Tree) MatchFirst(topic string) interface{} {
	values := t.Match(topic)

	if len(values) > 0 {
		return values[0]
	}

	return nil
}

// Search will return a set of values from topics that match the supplied topic.
// The result set will be cleared from duplicate values.
//
// Note: In contrast to Match, Search respects wildcards in the query but not in
// the stored tree.
func (t *Tree) Search(topic string) []interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	values := t.search([]interface{}{}, topic, t.root)

	return t.clean(values)
}

func (t *Tree) search(result []interface{}, topic string, node *node) []interface{} {
	// when finished add all values to the result set
	if topic == topicEnd {
		return append(result, node.values...)
	}

	// get segment
	segment := topicSegment(topic, t.Separator)

	// add all current and further values
	if segment == t.WildcardSome {
		result = append(result, node.values...)

		for _, child := range node.children {
			result = t.search(result, topic, child)
		}
	}

	// add all current values and continue
	if segment == t.WildcardOne {
		result = append(result, node.values...)

		for _, child := range node.children {
			result = t.search(result, topicShorten(topic, t.Separator), child)
		}
	}

	// match segments and get children
	if segment != t.WildcardOne && segment != t.WildcardSome {
		if child, ok := node.children[segment]; ok {
			result = t.search(result, topicShorten(topic, t.Separator), child)
		}
	}

	return result
}

// SearchFirst will run Search and return the first value or nil.
func (t *Tree) SearchFirst(topic string) interface{} {
	values := t.Search(topic)

	if len(values) > 0 {
		return values[0]
	}

	return nil
}

// clean will remove duplicates
func (t *Tree) clean(values []interface{}) []interface{} {
	result := values[:0]

	for _, v := range values {
		if contains(result, v) {
			continue
		}

		result = append(result, v)
	}

	return result
}

// Count will count all stored values in the tree. It will not filter out
// duplicate values and thus might return a different result to `len(All())`.
func (t *Tree) Count() int {
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
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.clean(t.all([]interface{}{}, t.root))
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
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.root = newNode()
}

// String will return a string representation of the tree.
func (t *Tree) String() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return fmt.Sprintf("topic.Tree:%s", t.root.string(0))
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

func topicShorten(topic, separator string) string {
	i := strings.Index(topic, separator)
	if i >= 0 {
		return topic[i+1:]
	}

	return topicEnd
}

func topicSegment(topic, separator string) string {
	i := strings.Index(topic, separator)
	if i >= 0 {
		return topic[:i]
	}

	return topic
}
