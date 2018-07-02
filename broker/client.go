package broker

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/transport"

	"gopkg.in/tomb.v2"
)

const (
	clientConnecting uint32 = iota
	clientConnected
	clientDisconnected
)

// ErrExpectedConnect is returned when the first received packet is not a
// ConnectPacket.
var ErrExpectedConnect = errors.New("expected a ConnectPacket as the first packet")

// ErrNotAuthorized is returned when a client is not authorized.
var ErrNotAuthorized = errors.New("client is not authorized")

// ErrMissingSession is returned if the backend does not return a session.
var ErrMissingSession = errors.New("no session returned from Backend")

// ErrDisconnected is returned if a client disconnects cleanly.
var ErrDisconnected = errors.New("disconnected")

// TODO: needed?

// A Client represents a remote client that is connected to the broker.
type Client struct {
	// Ref can be used to store a custom reference to an object. This is usually
	// used to attach a state object to client that is created in the Backend.
	Ref interface{}

	backend Backend
	logger  Logger
	conn    transport.Conn

	state        uint32
	clientID     string
	cleanSession bool
	session      Session

	inc chan packet.GenericPacket
	fwd chan *packet.Message

	tomb   tomb.Tomb
	mutex  sync.Mutex
	finish sync.Once
}

// NewClient takes over a connection and returns a Client.
func NewClient(backend Backend, logger Logger, conn transport.Conn) *Client {
	// create client
	c := &Client{
		state:   clientConnecting,
		backend: backend,
		logger:  logger,
		conn:    conn,
		inc:     make(chan packet.GenericPacket),
		fwd:     make(chan *packet.Message),
	}

	// start processor
	c.tomb.Go(c.processor)

	// run cleanup goroutine
	go func() {
		// wait for death and cleanup
		c.tomb.Wait()
		c.cleanup()
	}()

	return c
}

// Session returns the current Session used by the client.
func (c *Client) Session() Session {
	return c.session
}

// CleanSession returns whether the client requested a clean session during
// connect.
func (c *Client) CleanSession() bool {
	return c.cleanSession
}

// ClientID returns the supplied client id during connect.
func (c *Client) ClientID() string {
	return c.clientID
}

// RemoteAddr returns the client's remote net address from the underlying
// connection.
func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

/* goroutines */

// main processor
func (c *Client) processor() error {
	c.log(NewConnection, c, nil, nil, nil)

	// start receiver
	c.tomb.Go(c.receiver)

	// prepare packet
	var pkt packet.GenericPacket

	// get first packet from connection
	select {
	case pkt = <-c.inc:
		// continue
	case <-c.tomb.Dying():
		return tomb.ErrDying
	}

	// get connect
	connect, ok := pkt.(*packet.ConnectPacket)
	if !ok {
		return c.die(ClientError, ErrExpectedConnect)
	}

	// process connect
	err := c.processConnect(connect)
	if err != nil {
		return err // error has already been cleaned
	}

	// start requester
	c.tomb.Go(c.requester)

	for {
		select {
		case msg := <-c.fwd:
			// forward message
			err = c.forwardMessage(msg)
			if err != nil {
				return err // error has already been cleaned
			}
		case pkt = <-c.inc:
			// process packet
			err = c.processPacket(pkt)
			if err != nil {
				return err // error has already been cleaned
			}
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}
	}
}

// packet receiver
func (c *Client) receiver() error {
	for {
		// check if dying
		if !c.tomb.Alive() {
			return tomb.ErrDying
		}

		// receive next packet
		pkt, err := c.conn.Receive()
		if err != nil {
			return c.die(TransportError, err)
		}

		// send packet
		c.inc <- pkt
	}
}

// message requester
func (c *Client) requester() error {
	// TODO: Remove goroutine by exposing channels?

	for {
		// request next message
		msg, err := c.backend.Receive(c, c.tomb.Dying())
		if err != nil {
			return c.die(BackendError, err)
		} else if msg == nil {
			return tomb.ErrDying
		}

		// forward message
		c.fwd <- msg
	}
}

/* packet handling */

// handle an incoming ConnackPacket
func (c *Client) processConnect(pkt *packet.ConnectPacket) error {
	c.log(PacketReceived, c, pkt, nil, nil)

	// set values
	c.cleanSession = pkt.CleanSession
	c.clientID = pkt.ClientID

	// authenticate
	ok, err := c.backend.Authenticate(c, pkt.Username, pkt.Password)
	if err != nil {
		return c.die(BackendError, err)
	}

	// prepare connack packet
	connack := packet.NewConnackPacket()
	connack.ReturnCode = packet.ConnectionAccepted
	connack.SessionPresent = false

	// check authentication
	if !ok {
		// set return code
		connack.ReturnCode = packet.ErrNotAuthorized

		// send connack
		err = c.send(connack, false)
		if err != nil {
			return c.die(TransportError, err)
		}

		// close client
		return c.die(ClientError, ErrNotAuthorized)
	}

	// set state
	atomic.StoreUint32(&c.state, clientConnected)

	// set keep alive
	if pkt.KeepAlive > 0 {
		c.conn.SetReadTimeout(time.Duration(pkt.KeepAlive) * 1500 * time.Millisecond)
	} else {
		c.conn.SetReadTimeout(0)
	}

	// retrieve session
	s, resumed, err := c.backend.Setup(c, pkt.ClientID)
	if err != nil {
		return c.die(BackendError, err)
	} else if s == nil {
		return c.die(BackendError, ErrMissingSession)
	}

	// set state

	// reset the session if clean is requested
	if pkt.CleanSession {
		s.Reset()
		// TODO: Needed?
	}

	// set session present
	connack.SessionPresent = !pkt.CleanSession && resumed

	// assign session
	c.session = s

	// save will if present
	if pkt.Will != nil {
		err = c.session.SaveWill(pkt.Will)
		if err != nil {
			return c.die(SessionError, err)
		}
	}

	// send connack
	err = c.send(connack, false)
	if err != nil {
		return c.die(TransportError, err)
	}

	// retrieve stored packets
	packets, err := c.session.AllPackets(session.Outgoing)
	if err != nil {
		return c.die(SessionError, err)
	}

	// resend stored packets
	for _, pkt := range packets {
		// set the dup flag on a publish packet
		publish, ok := pkt.(*packet.PublishPacket)
		if ok {
			publish.Dup = true
		}

		// send packet
		err = c.send(pkt, true)
		if err != nil {
			return c.die(TransportError, err)
		}
	}

	// attempt to restore client if not clean
	if !pkt.CleanSession {
		// get stored subscriptions
		subs, err := s.AllSubscriptions()
		if err != nil {
			return c.die(SessionError, err)
		}

		// resubscribe subscriptions
		for _, sub := range subs {
			err = c.backend.Subscribe(c, sub)
			if err != nil {
				return c.die(BackendError, err)
			}
		}

		// begin with queueing offline messages
		err = c.backend.QueueOffline(c)
		if err != nil {
			return c.die(BackendError, err)
		}
	}

	return nil
}

// handle an incoming GenericPacket
func (c *Client) processPacket(pkt packet.GenericPacket) error {
	c.log(PacketReceived, c, pkt, nil, nil)

	// prepare error
	var err error

	// handle individual packets
	switch typedPkt := pkt.(type) {
	case *packet.SubscribePacket:
		err = c.processSubscribe(typedPkt)
	case *packet.UnsubscribePacket:
		err = c.processUnsubscribe(typedPkt)
	case *packet.PublishPacket:
		err = c.processPublish(typedPkt)
	case *packet.PubackPacket:
		err = c.processPubackAndPubcomp(typedPkt.ID)
	case *packet.PubcompPacket:
		err = c.processPubackAndPubcomp(typedPkt.ID)
	case *packet.PubrecPacket:
		err = c.processPubrec(typedPkt.ID)
	case *packet.PubrelPacket:
		err = c.processPubrel(typedPkt.ID)
	case *packet.PingreqPacket:
		err = c.processPingreq()
	case *packet.DisconnectPacket:
		err = c.processDisconnect()
	}

	// return eventual error
	if err != nil {
		return err // error has already been cleaned
	}

	return nil
}

// handle an incoming PingreqPacket
func (c *Client) processPingreq() error {
	// send a pingresp packet
	err := c.send(packet.NewPingrespPacket(), true)
	if err != nil {
		return c.die(TransportError, err)
	}

	return nil
}

// handle an incoming SubscribePacket
func (c *Client) processSubscribe(pkt *packet.SubscribePacket) error {
	// prepare suback packet
	suback := packet.NewSubackPacket()
	suback.ReturnCodes = make([]byte, len(pkt.Subscriptions))
	suback.ID = pkt.ID

	// handle contained subscriptions
	for i, subscription := range pkt.Subscriptions {
		// save subscription in session
		err := c.session.SaveSubscription(&subscription)
		if err != nil {
			return c.die(SessionError, err)
		}

		// subscribe client to queue
		err = c.backend.Subscribe(c, &subscription)
		if err != nil {
			return c.die(BackendError, err)
		}

		// save granted qos
		suback.ReturnCodes[i] = subscription.QOS
	}

	// send suback
	err := c.send(suback, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	// queue retained messages
	for _, sub := range pkt.Subscriptions {
		err := c.backend.QueueRetained(c, sub.Topic)
		if err != nil {
			return c.die(BackendError, err)
		}
	}

	return nil
}

// handle an incoming UnsubscribePacket
func (c *Client) processUnsubscribe(pkt *packet.UnsubscribePacket) error {
	// handle contained topics
	for _, topic := range pkt.Topics {
		// unsubscribe client from queue
		err := c.backend.Unsubscribe(c, topic)
		if err != nil {
			return c.die(BackendError, err)
		}

		// remove subscription from session
		err = c.session.DeleteSubscription(topic)
		if err != nil {
			return c.die(SessionError, err)
		}
	}

	// prepare unsuback packet
	unsuback := packet.NewUnsubackPacket()
	unsuback.ID = pkt.ID

	// send packet
	err := c.send(unsuback, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	return nil
}

// handle an incoming PublishPacket
func (c *Client) processPublish(publish *packet.PublishPacket) error {
	// handle unacknowledged and directly acknowledged messages
	if publish.Message.QOS <= 1 {
		err := c.handleMessage(&publish.Message)
		if err != nil {
			return c.die(BackendError, err)
		}
	}

	// handle qos 1 flow
	if publish.Message.QOS == 1 {
		puback := packet.NewPubackPacket()
		puback.ID = publish.ID

		// acknowledge qos 1 publish
		err := c.send(puback, true)
		if err != nil {
			return c.die(TransportError, err)
		}
	}

	// handle qos 2 flow
	if publish.Message.QOS == 2 {
		// store packet
		err := c.session.SavePacket(session.Incoming, publish)
		if err != nil {
			return c.die(SessionError, err)
		}

		// prepare pubrec packet
		pubrec := packet.NewPubrecPacket()
		pubrec.ID = publish.ID

		// signal qos 2 publish
		err = c.send(pubrec, true)
		if err != nil {
			return c.die(TransportError, err)
		}
	}

	return nil
}

// handle an incoming PubackPacket or PubcompPacket
func (c *Client) processPubackAndPubcomp(id packet.ID) error {
	// remove packet from store
	c.session.DeletePacket(session.Outgoing, id)

	return nil
}

// handle an incoming PubrecPacket
func (c *Client) processPubrec(id packet.ID) error {
	// allocate packet
	pubrel := packet.NewPubrelPacket()
	pubrel.ID = id

	// overwrite stored PublishPacket with PubrelPacket
	err := c.session.SavePacket(session.Outgoing, pubrel)
	if err != nil {
		return c.die(SessionError, err)
	}

	// send packet
	err = c.send(pubrel, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	return nil
}

// handle an incoming PubrelPacket
func (c *Client) processPubrel(id packet.ID) error {
	// get packet from store
	pkt, err := c.session.LookupPacket(session.Incoming, id)
	if err != nil {
		return c.die(SessionError, err)
	}

	// get packet from store
	publish, ok := pkt.(*packet.PublishPacket)
	if !ok {
		return nil // ignore a wrongly sent PubrelPacket
	}

	// publish packet to others
	err = c.handleMessage(&publish.Message)
	if err != nil {
		return c.die(BackendError, err)
	}

	// prepare pubcomp packet
	pubcomp := packet.NewPubcompPacket()
	pubcomp.ID = publish.ID

	// acknowledge PublishPacket
	err = c.send(pubcomp, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	// remove packet from store
	err = c.session.DeletePacket(session.Incoming, id)
	if err != nil {
		return c.die(SessionError, err)
	}

	return nil
}

// handle an incoming DisconnectPacket
func (c *Client) processDisconnect() error {
	// clear will
	err := c.session.ClearWill()
	if err != nil {
		return c.die(SessionError, err)
	}

	// mark client as cleanly disconnected
	atomic.StoreUint32(&c.state, clientDisconnected)

	// close underlying connection (triggers cleanup)
	c.conn.Close()

	c.log(ClientDisconnected, c, nil, nil, nil)

	return ErrDisconnected
}

/* helpers */

// handle publish messages
func (c *Client) handleMessage(msg *packet.Message) error {
	// check retain flag
	if msg.Retain {
		if len(msg.Payload) > 0 {
			// retain message
			err := c.backend.StoreRetained(c, msg)
			if err != nil {
				return err
			}
		} else {
			// clear already retained message
			err := c.backend.ClearRetained(c, msg.Topic)
			if err != nil {
				return err
			}
		}
	}

	// reset an existing retain flag
	msg.Retain = false

	// publish message to others
	err := c.backend.Publish(c, msg)
	if err != nil {
		return err
	}

	c.log(MessagePublished, c, nil, msg, nil)

	return nil
}

// forward messages
func (c *Client) forwardMessage(msg *packet.Message) error {
	// prepare publish packet
	publish := packet.NewPublishPacket()
	publish.Message = *msg

	// get stored subscription
	sub, err := c.session.LookupSubscription(publish.Message.Topic)
	if err != nil {
		return c.die(SessionError, err)
	}

	// check subscription
	if sub != nil {
		// respect maximum qos
		if publish.Message.QOS > sub.QOS {
			publish.Message.QOS = sub.QOS
		}
	}

	// set packet id
	if publish.Message.QOS > 0 {
		publish.ID = c.session.NextID()
	}

	// store packet if at least qos 1
	if publish.Message.QOS > 0 {
		err := c.session.SavePacket(session.Outgoing, publish)
		if err != nil {
			return c.die(SessionError, err)
		}
	}

	// send packet
	err = c.send(publish, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	c.log(MessageForwarded, c, nil, msg, nil)

	return nil
}

// send a packet
func (c *Client) send(pkt packet.GenericPacket, buffered bool) error {
	var err error

	// send packet
	if buffered {
		err = c.conn.BufferedSend(pkt)
	} else {
		err = c.conn.Send(pkt)
	}

	// check error
	if err != nil {
		return err
	}

	c.log(PacketSent, c, pkt, nil, nil)

	return nil
}

/* error handling and logging */

// log a message
func (c *Client) log(event LogEvent, client *Client, pkt packet.GenericPacket, msg *packet.Message, err error) {
	if c.logger != nil {
		c.logger(event, client, pkt, msg, err)
	}
}

// used for closing and cleaning up from internal goroutines
func (c *Client) die(event LogEvent, err error) error {
	// report error
	c.log(event, c, nil, nil, err)

	// close connection if requested
	c.conn.Close()

	return err
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup() {
	// check session
	if c.session != nil && atomic.LoadUint32(&c.state) == clientConnected {
		// get will
		will, err := c.session.LookupWill()
		if err != nil {
			c.log(SessionError, c, nil, nil, err)
		}

		// publish will message
		if will != nil {
			err = c.handleMessage(will)
			if err != nil {
				c.log(BackendError, c, nil, nil, err)
			}
		}
	}

	// remove client from the queue
	if atomic.LoadUint32(&c.state) >= clientConnected {
		err := c.backend.Terminate(c)
		if err != nil {
			c.log(BackendError, c, nil, nil, err)
		}

		// reset the session if clean is requested
		if c.session != nil && c.cleanSession {
			err = c.session.Reset()
			if err != nil {
				c.log(SessionError, c, nil, nil, err)
			}
		}
	}

	c.log(LostConnection, c, nil, nil, nil)
}
