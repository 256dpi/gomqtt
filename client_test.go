package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"net/url"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnConnect(func(sessionPresent bool) {
		require.False(t, sessionPresent)

		done <- true
	})

	err := c.QuickConnect("mqtt://localhost:1883", "test1")
	require.NoError(t, err)

	<-done

	c.Disconnect()
}

func TestClientPublishSubscribe(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnMessage(func(topic string, payload []byte) {
		require.Equal(t, "test", topic)
		require.Equal(t, []byte("test"), payload)

		done <- true
	})

	err := c.QuickConnect("mqtt://localhost:1883", "test1")
	require.NoError(t, err)

	c.Subscribe("test", 0)
	c.Publish("test", []byte("test"), 0, false)

	<-done

	c.Disconnect()
}

func TestClientConnectError(t *testing.T) {
	c := NewClient()

	err := c.QuickConnect("mqtt://localhost:1234", "test2") // wrong port
	require.Error(t, err)
}

func TestClientErrorAuthentication(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnError(func(err error) {
		require.Error(t, err)

		done <- true
	})

	err := c.Connect(&Options{
		URL: &url.URL{
			Scheme: "mqtt",
			Host:   "localhost:1883",
		}, // missing clientID
	})
	require.NoError(t, err)

	<-done
}
