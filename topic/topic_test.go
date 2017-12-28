package topic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTopic(t *testing.T) {
	tests := map[string]string{
		"topic/hello":         "topic/hello",
		"topic//hello":        "topic/hello",
		"topic///hello":       "topic/hello",
		"/topic":              "/topic",
		"//topic":             "/topic",
		"///topic":            "/topic",
		"topic/":              "topic",
		"topic//":             "topic",
		"topic///":            "topic",
		"topic///cool//hello": "topic/cool/hello",
		"topic//cool///hello": "topic/cool/hello",
	}

	for str, result := range tests {
		str, err := Parse(str, true)
		assert.Equal(t, result, str)
		assert.NoError(t, err, str)
	}
}

func TestZeroLengthError(t *testing.T) {
	_, err := Parse("", true)
	assert.Equal(t, ErrZeroLength, err)

	_, err = Parse("/", true)
	assert.Equal(t, ErrZeroLength, err)

	_, err = Parse("//", true)
	assert.Equal(t, ErrZeroLength, err)
}

func TestTopicDisallowWildcards(t *testing.T) {
	tests := map[string]bool{
		"topic":            true,
		"topic/hello":      true,
		"topic/cool/hello": true,
		"+":                false,
		"#":                false,
		"topic/+":          false,
		"topic/#":          false,
	}

	for str, result := range tests {
		_, err := Parse(str, false)

		if result {
			assert.NoError(t, err, str)
		} else {
			assert.Error(t, err, str)
		}
	}
}

func TestTopicAllowWildcards(t *testing.T) {
	tests := map[string]bool{
		"topic":            true,
		"topic/hello":      true,
		"topic/cool/hello": true,
		"+":                true,
		"#":                true,
		"topic/+":          true,
		"topic/#":          true,
		"topic/+/hello":    true,
		"topic/cool/+":     true,
		"topic/cool/#":     true,
		"+/cool/#":         true,
		"+/+/#":            true,
		"":                 false,
		"++":               false,
		"##":               false,
		"#/+":              false,
		"#/#":              false,
	}

	for str, result := range tests {
		_, err := Parse(str, true)

		if result {
			assert.NoError(t, err, str)
		} else {
			assert.Error(t, err, str)
		}
	}
}

func TestTopicContainsWildcards(t *testing.T) {
	assert.True(t, ContainsWildcards("topic/+"))
	assert.True(t, ContainsWildcards("topic/#"))
	assert.False(t, ContainsWildcards("topic/hello"))
}
