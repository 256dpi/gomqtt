package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalLaunch(t *testing.T) {
	server, err := Launch("tcp://localhost:0")
	require.NoError(t, err)

	err = server.Close()
	assert.NoError(t, err)
}

func TestLauncherBadURL(t *testing.T) {
	conn, err := Launch("foo")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestLauncherUnsupportedProtocol(t *testing.T) {
	conn, err := Launch("foo://localhost")
	assert.Nil(t, conn)
	assert.Equal(t, ErrUnsupportedProtocol, err)
}
