package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"net/url"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnConnect(func(sessionPresent bool){
		require.False(t, sessionPresent)

		done <- true
	})

	err := c.QuickConnect("mqtt://localhost:1883", "test1")
	require.NoError(t, err)

	<-done
}

func TestClientConnectError(t *testing.T) {
	c := NewClient()

	err := c.QuickConnect("mqtt://localhost:1234", "test2")
	require.Error(t, err)
}

func TestClientErrorAuthentication(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnError(func(err error){
		require.Error(t, err)

		done <- true
	})

	err := c.Connect(&Options{
		URL: &url.URL{
			Scheme: "mqtt",
			Host: "localhost:1883",
		},
	})
	require.NoError(t, err)

	<-done
}
