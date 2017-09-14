package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", "bar")
	assert.Equal(t, "bar", ctx.Get("foo"))
	assert.Nil(t, ctx.Get("baz"))
}
