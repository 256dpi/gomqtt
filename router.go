package router

import (
	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"gopkg.in/tomb.v2"
)

// A Handler receives incoming messages.
type Handler func(msg *packet.Message, params map[string]string)

// A Router wraps a service and provides routing of incoming messages.
type Router struct {
	tree   *Tree
	routes []*Route
	srv    *client.Service
	ctrl   *tomb.Tomb
	ch     chan *packet.Message
}

// New will create and return a new Router.
func New() *Router {
	return &Router{
		tree: NewTree(),
	}
}

// Handle will register the passed handler for the given filter. A matching
// subscription will be automatically created and issued when the service comes
// online.
func (r *Router) Handle(filter string, handler Handler) {
	route := r.tree.Add(filter, handler)
	r.routes = append(r.routes, route)
}

// Start will start the routers underlying service.
func (r *Router) Start(opts *client.Options) {
	r.srv = client.NewService(1)
	r.srv.Online = r.online
	r.srv.Message = r.message
	r.srv.Offline = r.offline
	r.srv.Start(opts)
}

// Publish will use the underlying service to publish the message.
func (r *Router) Publish(topic string, payload []byte, qos byte, retain bool) {
	if r.srv == nil {
		panic("router not started")
	}

	r.srv.Publish(topic, payload, qos, retain)
}

// Stop will stop the routers underlying service.
func (r *Router) Stop() {
	r.srv.Stop(true)
	r.srv = nil
}

func (r *Router) online(_ bool) {
	r.ctrl = new(tomb.Tomb)
	r.ch = make(chan *packet.Message)

	r.ctrl.Go(r.router)
}

func (r *Router) message(msg *packet.Message) {
	r.ch <- msg
}

func (r *Router) offline() {
	close(r.ch)
	r.ctrl.Kill(nil)
	r.ctrl.Wait()
	r.ctrl = nil
}

func (r *Router) router() error {
	for _, route := range r.routes {
		err := r.srv.Subscribe(route.RealFilter, 0).Wait()
		if err != nil {
			return err
		}
	}

	select {
	case msg := <-r.ch:
		value, params := r.tree.Route(msg.Topic)
		if value != nil {
			value.(Handler)(msg, params)
		}
	case <-r.ctrl.Dying():
		return tomb.ErrDying
	}

	return nil
}
