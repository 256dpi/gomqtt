package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTypes(t *testing.T) {
	if CONNECT != 1 ||
		CONNACK != 2 ||
		PUBLISH != 3 ||
		PUBACK != 4 ||
		PUBREC != 5 ||
		PUBREL != 6 ||
		PUBCOMP != 7 ||
		SUBSCRIBE != 8 ||
		SUBACK != 9 ||
		UNSUBSCRIBE != 10 ||
		UNSUBACK != 11 ||
		PINGREQ != 12 ||
		PINGRESP != 13 ||
		DISCONNECT != 14 {

		t.Errorf("Types have invalid code")
	}
}

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

	for _, _t := range list {
		m, err := _t.New()
		assert.NotNil(t, m)
		assert.NoError(t, err)
	}
}
