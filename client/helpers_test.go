package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTracker(t *testing.T) {
	tracker := newTracker(10 * time.Millisecond)
	assert.False(t, tracker.pending())
	assert.True(t, tracker.window() > 0)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, tracker.window() <= 0)

	tracker.reset()
	assert.True(t, tracker.window() > 0)

	tracker.ping()
	assert.True(t, tracker.pending())

	tracker.pong()
	assert.False(t, tracker.pending())
}
