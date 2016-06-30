package broker

import (
	"github.com/gomqtt/packet"
	"github.com/gomqtt/tools"
	"github.com/gomqtt/transport"
	"github.com/satori/go.uuid"
)

// a fake client for testing backend implementations
type fakeClient struct {
	in  []*packet.Message
	ctx *Context
}

// returns a new fake client
func newFakeClient() *fakeClient {
	ctx := NewContext()
	ctx.Set("uuid", uuid.NewV1().String())

	return &fakeClient{
		ctx: ctx,
	}
}

// publish will append the message to the in slice
func (c *fakeClient) Publish(msg *packet.Message) bool {
	c.in = append(c.in, msg)
	return true
}

// does nothing atm
func (c *fakeClient) Close(clean bool) {}

// returns the context
func (c *fakeClient) Context() *Context {
	return c.ctx
}

// runs broker on a test port
func runBroker(broker *Broker, num int) (*tools.Port, chan struct{}) {
	port := tools.NewPort()

	server, _ := transport.Launch(port.URL())

	done := make(chan struct{})

	go func() {
		for i := 0; i < num; i++ {
			conn, _ := server.Accept()

			if conn != nil {
				broker.Handle(conn)
			}
		}
	}()

	go func() {
		<-done
		server.Close()
	}()

	return port, done
}
