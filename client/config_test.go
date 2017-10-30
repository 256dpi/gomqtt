package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {
	config := NewConfig("foo")
	assert.Equal(t, "foo", config.BrokerURL)
	assert.Equal(t, "", config.ClientID)
	assert.True(t, config.CleanSession)
	assert.Equal(t, "30s", config.KeepAlive)
}
