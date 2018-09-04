package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTracker(t *testing.T) {
	tracker := NewTracker(10 * time.Millisecond)
	assert.False(t, tracker.Pending())
	assert.True(t, tracker.Window() > 0)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, tracker.Window() <= 0)

	tracker.Reset()
	assert.True(t, tracker.Window() > 0)

	tracker.Ping()
	assert.True(t, tracker.Pending())

	tracker.Pong()
	assert.False(t, tracker.Pending())
}
