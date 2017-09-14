package router

import (
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
)

func Example() {
	r := New()

	done := make(chan struct{})

	r.Handle("device/+id/#sensor", Logger(func(w ResponseWriter, r *Request) {
		w.Publish(&packet.Message{
			Topic:   "finish/data",
			Payload: []byte("7"),
		})
	}))

	r.Handle("finish/data", Logger(func(w ResponseWriter, r *Request) {
		close(done)
	}))

	config := client.NewConfig("mqtt://try:try@broker.shiftr.io")

	r.Start(config)

	time.Sleep(2 * time.Second)

	r.Publish(&packet.Message{
		Topic:   "device/foo/bar/baz",
		Payload: []byte("42"),
	})

	<-done

	r.Stop()

	// Output:
	// New Request: <Message Topic="device/foo/bar/baz" QOS=0 Retain=false Payload=[52 50]>
	// Publishing: <Message Topic="finish/data" QOS=0 Retain=false Payload=[55]>
	// New Request: <Message Topic="finish/data" QOS=0 Retain=false Payload=[55]>
}
