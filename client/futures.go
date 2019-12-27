package client

import (
	"time"

	"github.com/256dpi/gomqtt/client/future"
	"github.com/256dpi/gomqtt/packet"
)

// A GenericFuture is returned by publish and unsubscribe methods.
type GenericFuture interface {
	// Wait will block until the future is completed or canceled. It will return
	// future.ErrCanceled if the future gets canceled. If the timeout is reached,
	// future.ErrTimeoutExceeded is returned.
	//
	// Note: Wait will not return any Client related errors.
	Wait(timeout time.Duration) error
}

// A ConnectFuture is returned by the connect method.
type ConnectFuture interface {
	GenericFuture

	// SessionPresent will return whether a session was present.
	SessionPresent() bool

	// ReturnCode will return the connack code returned by the broker.
	ReturnCode() packet.ConnackCode
}

// A SubscribeFuture is returned by the subscribe methods.
type SubscribeFuture interface {
	GenericFuture

	// ReturnCodes will return the suback codes returned by the broker.
	ReturnCodes() []packet.QOS
}

type connectFuture struct {
	*future.Future
}

func (f *connectFuture) SessionPresent() bool {
	// get result
	v, ok := f.Result().(*packet.Connack)
	if !ok {
		return false
	}

	return v.SessionPresent
}

func (f *connectFuture) ReturnCode() packet.ConnackCode {
	// get result
	v, ok := f.Result().(*packet.Connack)
	if !ok {
		return 0
	}

	return v.ReturnCode
}

type subscribeFuture struct {
	*future.Future
}

func (f *subscribeFuture) ReturnCodes() []packet.QOS {
	// get result
	v, ok := f.Result().(*packet.Suback)
	if !ok {
		return nil
	}

	return v.ReturnCodes
}
