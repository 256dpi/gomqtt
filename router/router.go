package router

import (
	"time"

	"github.com/gomqtt/client"
	"github.com/gomqtt/packet"
	"gopkg.in/tomb.v2"
)

// A Request holds the incoming routed message and the parsed params.
type Request struct {
	Message *packet.Message
	Params  map[string]string
}

// A ResponseWriter is used to publish messages to the broker.
type ResponseWriter interface {
	Publish(*packet.Message)
}

// A Handler receives incoming requests.
type Handler func(ResponseWriter, *Request)

// A Router wraps a service and provides routing of incoming messages.
type Router struct {
	SubscribeTimeout time.Duration

	tree   *Tree
	routes []*Route
	srv    *client.Service
	ctrl   *tomb.Tomb
	ch     chan *packet.Message
}

// New will create and return a new Router.
//
// See: client.NewService.
func New(queueSize ...int) *Router {
	r := &Router{
		SubscribeTimeout: 1 * time.Second,

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

// Publish will use the underlying service to publish the specified message.
func (r *Router) Publish(msg *packet.Message) {
	r.srv.PublishMessage(msg)
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
		err := r.srv.Subscribe(route.RealFilter, 0).Wait(r.SubscribeTimeout)
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
