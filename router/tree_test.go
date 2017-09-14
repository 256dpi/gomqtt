package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeRoute(t *testing.T) {
	tree := NewTree()

	route := tree.Add("foo/+p1/bar/#p2", 1)
	assert.Equal(t, &Route{
		Filter:     "foo/+p1/bar/#p2",
		Value:      1,
		RealFilter: "foo/+/bar/#",
		Params:     map[int]string{1: "p1"},
		SplatIndex: 3,
		SplatName:  "p2",
	}, route)

	value, params := tree.Route("foo/aaa/bar/bbb/ccc")
	assert.Equal(t, 1, value)
	assert.Equal(t, "aaa", params["p1"])
	assert.Equal(t, "bbb/ccc", params["p2"])
}
