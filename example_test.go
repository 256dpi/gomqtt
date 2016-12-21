package router

import (
	"fmt"
	"time"

	"github.com/gomqtt/client"
)

func Example() {
	r := New()

	done := make(chan struct{})

	r.Handle("device/+id/#sensor", func(w ResponseWriter, r *Request) {
		fmt.Println(r.Params["id"])
		fmt.Println(r.Params["sensor"])
		fmt.Println(string(r.Message.Payload))

		w.Publish("finish/data", []byte("7"), 0, false)
	})

	r.Handle("finish/data", func(w ResponseWriter, r *Request) {
		fmt.Println(string(r.Message.Payload))

		close(done)
	})

	config := client.NewConfig("mqtt://try:try@broker.shiftr.io")

	r.Start(config)

	time.Sleep(2 * time.Second)

	r.Publish("device/foo/bar/baz", []byte("42"), 0, false)

	<-done

	r.Stop()

	// Output:
	// foo
	// bar/baz
	// 42
	// 7
}
