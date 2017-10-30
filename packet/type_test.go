package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTypeString(t *testing.T) {
	assert.Equal(t, "Unknown", Type(99).String())
}

func TestTypeValid(t *testing.T) {
	assert.True(t, CONNECT.Valid())
}

func TestTypeNew(t *testing.T) {
	list := []Type{
		CONNECT,
		CONNACK,
		PUBLISH,
		PUBACK,
		PUBREC,
		PUBREL,
		PUBCOMP,
		SUBSCRIBE,
		SUBACK,
		UNSUBSCRIBE,
		UNSUBACK,
		PINGREQ,
		PINGRESP,
		DISCONNECT,
	}

	for _, tt := range list {
		m, err := tt.New()
		assert.NotNil(t, m)
		assert.NoError(t, err)
	}
}
