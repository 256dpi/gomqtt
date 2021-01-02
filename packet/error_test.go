package packet

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := wrapError(PUBLISH, DECODE, M4, 7, ErrInvalidTopic)
	assert.Equal(t, &Error{Type: PUBLISH, Method: DECODE, Mode: M4, Position: 7, Err: ErrInvalidTopic}, err)
	assert.Equal(t, "invalid topic", err.Error())
	assert.Equal(t, ErrInvalidTopic, errors.Unwrap(err))

	err = makeError(PUBLISH, DECODE, M4, 7, "foo")
	assert.Equal(t, &Error{Type: PUBLISH, Method: DECODE, Mode: M4, Position: 7, Message: "foo"}, err)
	assert.Equal(t, "foo", err.Error())
	assert.Nil(t, errors.Unwrap(err))
}
