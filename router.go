package router

import (
	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"gopkg.in/tomb.v2"
)

// A Request holds the incoming routed message and some meta data.
type Request struct {
	Message *packet.Message
	Params  map[string]string
}

// A ResponseWriter is used to convey messages to the broker.
type ResponseWriter interface {
	Publish(topic string, payload []byte, qos byte, retain bool)
}

// A Handler receives incoming requests.
type Handler func(ResponseWriter, *Request)

// A Router wraps a service and provides routing of incoming messages.
type Router struct {
	tree   *Tree
	routes []*Route
	srv    *client.Service
	ctrl   *tomb.Tomb
	ch     chan *packet.Message
}

// New will create and return a new Router.
func New(queueSize ...int) *Router {
	r := &Router{
		tree: NewTree(),
		srv:  client.NewService(queueSize...),
	}

	r.srv.OnlineCallback = r.onlineCallback
	r.srv.MessageCallback = r.messageCallback
	r.srv.OfflineCallback = r.offlineCallback

	return r
}

// Handle will register the passed handler for the given filter. A matching
// subscription will be automatically created and issued when the service comes
// online.
func (r *Router) Handle(filter string, handler Handler) {
	route := r.tree.Add(filter, handler)
	r.routes = append(r.routes, route)
}

// Start will start the routers underlying service.
func (r *Router) Start(opts *client.Config) {
	r.srv.Start(opts)
}

// Publish will use the underlying service to publish the message.
func (r *Router) Publish(topic string, payload []byte, qos byte, retain bool) {
	r.srv.Publish(topic, payload, qos, retain)
}

// Stop will stop the routers underlying service.
func (r *Router) Stop() {
	r.srv.Stop(true)
}

func (r *Router) onlineCallback(_ bool) {
	r.ctrl = new(tomb.Tomb)
	r.ch = make(chan *packet.Message)

	r.ctrl.Go(r.router)
}

func (r *Router) messageCallback(msg *packet.Message) {
	r.ch <- msg
}

func (r *Router) offlineCallback() {
	r.ctrl.Kill(nil)
	r.ctrl.Wait()

	close(r.ch)
	r.ctrl = nil
}

func (r *Router) router() error {
	for _, route := range r.routes {
		// TODO: Use timeout.
		err := r.srv.Subscribe(route.RealFilter, 0).Wait()
		if err != nil {
			return err
		}
	}

	for {
		select {
		case msg := <-r.ch:
			value, params := r.tree.Route(msg.Topic)
			if value != nil {
				value.(Handler)(r, &Request{
					Message: msg,
					Params:  params,
				})
			}
		case <-r.ctrl.Dying():
			return tomb.ErrDying
		}
	}
}
