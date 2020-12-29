package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	assert.Len(t, Types(), 15)
}

func TestTypeString(t *testing.T) {
	for _, tt := range Types() {
		assert.True(t, tt.String() != "Unknown" && tt.String() != "")
	}

	assert.Equal(t, "Unknown", Type(99).String())
}

func TestTypeValid(t *testing.T) {
	for _, tt := range Types() {
		assert.True(t, tt.Valid())
	}
}

func TestTypeNew(t *testing.T) {
	for _, tt := range Types() {
		m, err := tt.New()
		assert.NotNil(t, m)
		assert.NoError(t, err)
	}
}
