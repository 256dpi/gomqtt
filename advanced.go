package spec

import (
	"testing"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func AuthenticationTest(t *testing.T, url, denyURL string) {
	// client1 should be denied

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		assert.Equal(t, client.ErrClientConnectionDenied, err)
	}

	connectFuture1, err := client1.Connect(denyURL, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ErrNotAuthorized, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	// client2 should be allowed

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, nil)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	err = client2.Disconnect()
	assert.NoError(t, err)
}

func UniqueClientIDTest(t *testing.T, url string, id string) {
	assert.NoError(t, client.ClearSession(url, id))

	options := client.NewOptions()
	options.ClientID = id

	wait := make(chan struct{})

	/* first client */

	client1 := client.New()
	client1.Callback = func(msg *packet.Message, err error) {
		assert.Error(t, err)
		close(wait)
	}

	connectFuture1, err := client1.Connect(url, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture1.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture1.ReturnCode)
	assert.False(t, connectFuture1.SessionPresent)

	/* second client */

	client2 := client.New()

	connectFuture2, err := client2.Connect(url, options)
	assert.NoError(t, err)
	assert.NoError(t, connectFuture2.Wait())
	assert.Equal(t, packet.ConnectionAccepted, connectFuture2.ReturnCode)
	assert.False(t, connectFuture2.SessionPresent)

	<-wait

	err = client2.Disconnect()
	assert.NoError(t, err)
}
