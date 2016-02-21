package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gomqtt/packet"
)

// BackendSpec will test a Backend implementation. The test will test all methods
// using a fake consumer. The passed builder callback should always return a
// fresh instances of the Backend.
//
// Authentication: Its expected that the Backend allows the login "allow:allow".
func BackendSpec(t *testing.T, builder func() Backend) {
	t.Log("Running Backend Authentication Test")
	backendAuthenticationTest(t, builder())

	t.Log("Running Backend Setup Test")
	backendSetupTest(t, builder())

	t.Log("Running Backend Basic Queuing Test")
	backendBasicQueuingTest(t, builder())

	// TODO: Offline Message Test?

	t.Log("Running Backend Retained Messages Test")
	backendRetainedMessagesTest(t, builder())
}

func backendAuthenticationTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	ok, err := backend.Authenticate(consumer, "allow", "allow")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = backend.Authenticate(consumer, "deny", "deny")
	assert.False(t, ok)
	assert.NoError(t, err)
}

func backendSetupTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	// has id and clean=false

	session1, resumed, err := backend.Setup(consumer, "foo", false)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session1)

	session2, resumed, err := backend.Setup(consumer, "foo", false)
	assert.NoError(t, err)
	assert.True(t, resumed)
	assert.True(t, session1 == session2)

	// has id and clean=true

	session3, resumed, err := backend.Setup(consumer, "foo", true)
	assert.NoError(t, err)
	assert.True(t, resumed)
	assert.True(t, session1 == session3)

	// has other id and clean=false

	session4, resumed, err := backend.Setup(consumer, "bar", false)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.True(t, session4 != session1)
	assert.True(t, session4 != session2)
	assert.True(t, session4 != session3)

	// has no id and clean=true

	session5, resumed, err := backend.Setup(consumer, "", true)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session5)

	session6, resumed, err := backend.Setup(consumer, "", true)
	assert.NoError(t, err)
	assert.False(t, resumed)
	assert.NotNil(t, session5)
	assert.True(t, session5 != session6)
}

func backendBasicQueuingTest(t *testing.T, backend Backend) {
	consumer1 := newFakeConsumer()
	consumer2 := newFakeConsumer()

	msg := &packet.Message{
		Topic:   "test",
		Payload: []byte("test"),
	}

	// setup both consumers

	_, _, err := backend.Setup(consumer1, "consumer1", true)
	assert.NoError(t, err)

	_, _, err = backend.Setup(consumer2, "consumer2", true)
	assert.NoError(t, err)

	// subscribe both consumers

	msgs, err := backend.Subscribe(consumer1, "test")
	assert.Nil(t, msgs)
	assert.NoError(t, err)

	msgs, err = backend.Subscribe(consumer2, "test")
	assert.Nil(t, msgs)
	assert.NoError(t, err)

	// publish message

	err = backend.Publish(consumer1, msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, consumer1.in[0])
	assert.Equal(t, msg, consumer2.in[0])

	// unsubscribe one consumer

	err = backend.Unsubscribe(consumer2, "test")
	assert.NoError(t, err)

	// publish another message

	err = backend.Publish(consumer1, msg)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(consumer1.in))
	assert.Equal(t, 1, len(consumer2.in))

	// terminate both consumers

	err = backend.Terminate(consumer1)
	assert.NoError(t, err)

	err = backend.Terminate(consumer2)
	assert.NoError(t, err)
}

func backendRetainedMessagesTest(t *testing.T, backend Backend) {
	consumer := newFakeConsumer()

	msg1 := &packet.Message{
		Topic:   "foo",
		Payload: []byte("bar"),
		QOS:     1,
		Retain:  true,
	}

	msg2 := &packet.Message{
		Topic:   "foo/bar",
		Payload: []byte("bar"),
		QOS:     1,
		Retain:  true,
	}

	msg3 := &packet.Message{
		Topic:   "foo",
		Payload: []byte("bar"),
		QOS:     2,
		Retain:  true,
	}

	msg4 := &packet.Message{
		Topic:  "foo",
		QOS:    1,
		Retain: true,
	}

	// should be empty
	msgs, err := backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)

	err = backend.Publish(consumer, msg1)
	assert.NoError(t, err)

	// should have one
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg1, msgs[0])

	err = backend.Publish(consumer, msg2)
	assert.NoError(t, err)

	// should have two
	msgs, err = backend.Subscribe(consumer, "#")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(msgs))

	err = backend.Publish(consumer, msg3)
	assert.NoError(t, err)

	// should have another
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Equal(t, msg3, msgs[0])

	err = backend.Publish(consumer, msg4)
	assert.NoError(t, err)

	// should have none
	msgs, err = backend.Subscribe(consumer, "foo")
	assert.NoError(t, err)
	assert.Empty(t, msgs)
}
