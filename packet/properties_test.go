package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProperties(t *testing.T) {
	for c := range propertyNames {
		assert.NotZero(t, propertyTypes[c])
	}

	for c := range propertyTypes {
		assert.NotEmpty(t, t, propertyNames[c])
	}
}
