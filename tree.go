package router

import (
	"strings"

	"github.com/gomqtt/tools"
)

// A Route is single route added to the Tree.
type Route struct {
	Filter     string
	Value      interface{}
	RealFilter string
	Params     map[int]string
	SplatIndex int
	SplatName  string
}

// A Tree extends the basic tools.Tree to handle named parameters.
type Tree struct {
	tree *tools.Tree
}

// NewTree creates and returns a new Tree.
func NewTree() *Tree {
	return &Tree{
		tree: tools.NewTree(),
	}
}

// Add registers the value for the supplied topic. The function will parse the
// topic into a Route.
func (t *Tree) Add(filter string, value interface{}) *Route {
	route := &Route{
		Filter:     filter,
		Value:      value,
		Params:     make(map[int]string, 0),
		SplatIndex: -1,
	}

	segments := strings.Split(filter, "/")
	realFilter := []string{}

	for i, segment := range segments {
		isParam := strings.HasPrefix(segment, "+")
		isSplat := strings.HasPrefix(segment, "#")

		if !isParam && !isSplat {
			realFilter = append(realFilter, segment)
			continue
		}

		if len(segment) <= 1 {
			panic("supplied param or splat must have a name")
		}

		if isParam {
			route.Params[i] = segment[1:]
			realFilter = append(realFilter, "+")
		} else if isSplat {
			if route.SplatIndex >= 0 {
				panic("only one splat param is allowed")
			}

			route.SplatIndex = i
			route.SplatName = segment[1:]
			realFilter = append(realFilter, "#")
		}
	}

	route.RealFilter = strings.Join(realFilter, "/")

	t.tree.Add(route.RealFilter, route)

	return route
}

// Route will find and return the matching Route for the supplied topic. The
// function will also return a map with the parsed params.
func (t *Tree) Route(topic string) (interface{}, map[string]string) {
	route := t.tree.MatchFirst(topic).(*Route)
	if route == nil {
		return nil, nil
	}

	params := make(map[string]string, 0)

	segments := strings.Split(topic, "/")
	for index, name := range route.Params {
		params[name] = segments[index]
	}

	if route.SplatIndex >= 0 {
		params[route.SplatName] = strings.Join(segments[route.SplatIndex:], "/")
	}

	return route.Value, params
}
