package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	c := NewClient()

	done := make(chan bool)

	c.OnConnect(func(sessionPresent bool){
		require.False(t, sessionPresent)

		done <- true
	})

	c.Connect("ignored")

	<-done
}
