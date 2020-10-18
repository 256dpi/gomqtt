package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	assert.Len(t, Types(), 14)
}

func TestTypeString(t *testing.T) {
	assert.Equal(t, "Unknown", Type(99).String())
}

func TestTypeValid(t *testing.T) {
	assert.True(t, CONNECT.Valid())
}

func TestTypeNew(t *testing.T) {
	for _, tt := range Types() {
		m, err := tt.New()
		assert.NotNil(t, m)
		assert.NoError(t, err)
	}
}
