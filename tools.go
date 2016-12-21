package router

import (
	"fmt"

	"github.com/gomqtt/packet"
)

type logWriter struct {
	w ResponseWriter
}

func (w *logWriter) Publish(msg *packet.Message) {
	fmt.Printf("Publishing: %s\n", msg)

	w.w.Publish(msg)
}

// Logger is a middleware that prints requests and published messages.
func Logger(next Handler) Handler {
	return func(w ResponseWriter, r *Request) {
		// TODO: Sort map params.
		fmt.Printf("New Request: %s\n", r)

		next(&logWriter{w}, r)
	}
}
