package router

import (
	"fmt"
	"time"

	"github.com/gomqtt/packet"
	"github.com/gomqtt/client"
)

func Example() {
	r := New()

	done := make(chan struct{})

	r.Handle("device/+id/#sensor", func(msg *packet.Message, params map[string]string) {
		fmt.Println(params["id"])
		fmt.Println(params["sensor"])
		fmt.Println(string(msg.Payload))

		close(done)
	})

	r.Start(client.NewOptions("mqtt://localhost"))

	time.Sleep(100 * time.Millisecond)

	r.Publish("device/foo/bar/baz", []byte("42"), 0, false)

	<-done

	r.Stop()

	// Output:
	// foo
	// bar/baz
	// 42
}
