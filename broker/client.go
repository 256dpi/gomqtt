package broker

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/transport"

	"gopkg.in/tomb.v2"
)

// LogEvent denotes the class of an event passed to the logger.
type LogEvent string

const (
	// NewConnection is emitted when a client comes online.
	NewConnection LogEvent = "new connection"

	// PacketReceived is emitted when a packet has been received.
	PacketReceived LogEvent = "packet received"

	// MessagePublished is emitted after a message has been published.
	MessagePublished LogEvent = "message published"

	// MessageAcknowledged is emitted after a message has been acknowledged.
	MessageAcknowledged LogEvent = "message acknowledged"

	// MessageDequeued is emitted after a message has been dequeued.
	MessageDequeued LogEvent = "message dequeued"

	// MessageForwarded is emitted after a message has been forwarded.
	MessageForwarded LogEvent = "message forwarded"

	// PacketSent is emitted when a packet has been sent.
	PacketSent LogEvent = "packet sent"

	// ClientDisconnected is emitted when a client disconnects cleanly.
	ClientDisconnected LogEvent = "client disconnected"

	// TransportError is emitted when an underlying transport error occurs.
	TransportError LogEvent = "transport error"

	// SessionError is emitted when a call to the session fails.
	SessionError LogEvent = "session error"

	// BackendError is emitted when a call to the backend fails.
	BackendError LogEvent = "backend error"

	// ClientError is emitted when the client violates the protocol.
	ClientError LogEvent = "client error"

	// LostConnection is emitted when the connection has been terminated.
	LostConnection LogEvent = "lost connection"
)

// A Session is used to get packet ids and persist incoming/outgoing packets.
type Session interface {
	// NextID should return the next id for outgoing packets.
	NextID() packet.ID

	// SavePacket should store a packet in the session. An eventual existing
	// packet with the same id should be quietly overwritten.
	SavePacket(session.Direction, packet.Generic) error

	// LookupPacket should retrieve a packet from the session using the packet id.
	LookupPacket(session.Direction, packet.ID) (packet.Generic, error)

	// DeletePacket should remove a packet from the session. The method should
	// not return an error if no packet with the specified id does exists.
	DeletePacket(session.Direction, packet.ID) error

	// AllPackets should return all packets currently saved in the session.
	AllPackets(session.Direction) ([]packet.Generic, error)
}

// Ack is executed by the Backend or Client to signal either that a message will
// be delivered under the selected qos level and is therefore safe to be deleted
// from either queue or the successful handling of subscriptions.
type Ack func()

// A Backend provides the effective brokering functionality to its clients.
type Backend interface {
	// Authenticate should authenticate the client using the user and password
	// values and return true if the client is eligible to continue or false
	// when the broker should terminate the connection.
	Authenticate(client *Client, user, password string) (ok bool, err error)

	// Setup is called when a new client comes online and is successfully
	// authenticated. Setup should return the already stored session for the
	// supplied id or create and return a new one if it is missing or a clean
	// session is requested. If the supplied id has a zero length, a new
	// temporary session should returned that is not stored further. The backend
	// should also close any existing clients that use the same client id.
	//
	// Note: In this call the Backend may also allocate other resources and
	// setup the client for further usage as the broker will acknowledge the
	// connection when the call returns. The Terminate function is called for
	// every client that Setup has been called for.
	Setup(client *Client, id string, clean bool) (a Session, resumed bool, err error)

	// Restore is called after the client has restored packets from the session.
	//
	// The Backend should resubscribe stored subscriptions. Retained messages
	// that are delivered as a result of resubscribing a stored subscription
	// must be queued with the retain flag set to false.
	Restore(client *Client) error

	// Subscribe should subscribe the passed client to the specified topics and
	// store the subscription in the session. If an Ack is provided, the
	// subscription will be acknowledged when called during or after the call to
	// Subscribe.
	//
	// Incoming messages that match the supplied subscription should be added to
	// a temporary or persistent queue that is drained when Dequeue is called.
	//
	// Retained messages that match the supplied subscription should be added to
	// a temporary queue that is also drained when Dequeue is called. Retained
	// messages are not part of the stored session state as they are anyway
	// redelivered using the stored subscription mechanism.
	Subscribe(client *Client, subs []packet.Subscription, ack Ack) error

	// Unsubscribe should unsubscribe the passed client from the specified topics
	// and remove the subscriptions from the session. If an Ack is provided, the
	// unsubscription will be acknowledged when called during or after the call
	// to Unsubscribe.
	Unsubscribe(client *Client, topics []string, ack Ack) error

	// Publish should forward the passed message to all other clients that hold
	// a subscription that matches the messages topic. It should also add the
	// message to all sessions that have a matching offline subscription. The
	// later may only apply to messages with a QoS greater than 0. If an Ack is
	// provided, the message will be acknowledged when called during or after
	// the call to Publish.
	//
	// If the retained flag is set, messages with a payload should replace the
	// currently retained message. Otherwise, the currently retained message
	// should be removed. The flag should be cleared before publishing the
	// message to subscribed clients.
	Publish(client *Client, msg *packet.Message, ack Ack) error

	// Dequeue is called by the Client to obtain the next message from the queue
	// and must return either a message or an error. The backend must only return
	// no message and no error if the client's Closing channel has been closed.
	//
	// The Backend may return an Ack to receive a signal that the message is being
	// delivered under the selected qos level and is therefore safe to be deleted
	// from the queue. The Client might dequeue other messages before acknowledging
	// a message.
	//
	// The returned message must have a QoS set that respects the QoS set by
	// the matching subscription.
	Dequeue(client *Client) (*packet.Message, Ack, error)

	// Terminate is called when the client goes offline. Terminate should
	// unsubscribe the passed client from all previously subscribed topics. The
	// backend may also convert a clients subscriptions to offline subscriptions.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that client as the broker will close the connection when the call
	// returns.
	Terminate(client *Client) error

	// Log is called multiple times during the lifecycle of a client see LogEvent
	// for a list of all events.
	Log(event LogEvent, client *Client, pkt packet.Generic, msg *packet.Message, err error)
}

// ErrExpectedConnect is returned when the first received packet is not a
// Connect.
var ErrExpectedConnect = errors.New("expected a Connect as the first packet")

// ErrNotAuthorized is returned when a client is not authorized.
var ErrNotAuthorized = errors.New("client is not authorized")

// ErrMissingSession is returned if the backend does not return a session.
var ErrMissingSession = errors.New("no session returned from Backend")

// ErrClientDisconnected is returned if a client disconnects cleanly.
var ErrClientDisconnected = errors.New("client has disconnected")

// ErrClientClosed is returned if a client is being closed by the broker.
var ErrClientClosed = errors.New("client has been closed")

const (
	clientConnecting uint32 = iota
	clientConnected
	clientDisconnected
)

type outgoing struct {
	pkt packet.Generic
	msg *packet.Message
	ack Ack
}

// A Client represents a remote client that is connected to the broker.
type Client struct {
	// PacketPrefetch may be set during Setup to control the number of packets
	// that are read by Client and made available for processing. Will default
	// to 10 if not set.
	PacketPrefetch int

	// ParallelPublishes may be set during Setup to control the number of
	// parallel calls to Publish a client can perform. Will default to 10.
	ParallelPublishes int

	// ParallelSubscribes may be set during Setup to control the number of
	// parallel calls to Subscribe and Unsubscribe a client can perform.
	// Will default to 10.
	ParallelSubscribes int

	// ParallelDequeues may be set during Setup to control the number of
	// parallel calls to Dequeue a client can perform. Will default to 10.
	ParallelDequeues int

	// read-only
	backend Backend
	conn    transport.Conn

	// atomically written and read
	state uint32

	// set during connect
	id      string
	will    *packet.Message
	session Session

	incoming chan packet.Generic
	outgoing chan outgoing

	publishTokens   chan struct{}
	subscribeTokens chan struct{}
	dequeueTokens   chan struct{}

	tomb tomb.Tomb
	done chan struct{}
}

// NewClient takes over a connection and returns a Client.
func NewClient(backend Backend, conn transport.Conn) *Client {
	// create client
	c := &Client{
		state:   clientConnecting,
		backend: backend,
		conn:    conn,
		done:    make(chan struct{}),
	}

	// start processor
	c.tomb.Go(c.processor)

	// run cleanup goroutine
	go func() {
		// wait for death and cleanup
		c.tomb.Wait()
		c.cleanup()

		// close channel
		close(c.done)
	}()

	return c
}

// Session returns the current Session used by the client.
func (c *Client) Session() Session {
	return c.session
}

// ID returns the clients id that has been supplied during connect.
func (c *Client) ID() string {
	return c.id
}

// Conn returns the client's underlying connection. Calls to SetReadLimit,
// SetBuffers, LocalAddr and RemoteAddr are safe.
func (c *Client) Conn() transport.Conn {
	return c.conn
}

// Close will immediately close the client.
func (c *Client) Close() {
	c.tomb.Kill(ErrClientClosed)
	c.conn.Close()
}

// Closing returns a channel that is closed when the client is closing.
func (c *Client) Closing() <-chan struct{} {
	return c.tomb.Dying()
}

// Closed returns a channel that is closed when the client is closed.
func (c *Client) Closed() <-chan struct{} {
	return c.done
}

/* goroutines */

// main processor
func (c *Client) processor() error {
	c.backend.Log(NewConnection, c, nil, nil, nil)

	// get first packet from connection
	pkt, err := c.conn.Receive()
	if err != nil {
		return c.die(TransportError, err)
	}

	c.backend.Log(PacketReceived, c, pkt, nil, nil)

	// get connect
	connect, ok := pkt.(*packet.Connect)
	if !ok {
		return c.die(ClientError, ErrExpectedConnect)
	}

	// process connect
	err = c.processConnect(connect)
	if err != nil {
		return err // error has already been cleaned
	}

	// start dequeuer and sender
	c.tomb.Go(c.dequeuer)
	c.tomb.Go(c.sender)

	for {
		// check if still alive
		if !c.tomb.Alive() {
			return tomb.ErrDying
		}

		// receive next packet
		pkt, err := c.conn.Receive()
		if err != nil {
			return c.die(TransportError, err)
		}

		c.backend.Log(PacketReceived, c, pkt, nil, nil)

		// process packet
		err = c.processPacket(pkt)
		if err != nil {
			return err // error has already been cleaned
		}
	}
}

// message dequeuer
func (c *Client) dequeuer() error {
	for {
		// acquire dequeue token
		select {
		case <-c.dequeueTokens:
			// continue
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}

		// request next message
		msg, ack, err := c.backend.Dequeue(c)
		if err != nil {
			return c.die(BackendError, err)
		} else if msg == nil {
			return tomb.ErrDying
		}

		c.backend.Log(MessageDequeued, c, nil, msg, nil)

		// queue message
		select {
		case c.outgoing <- outgoing{msg: msg, ack: ack}:
			// continue
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}
	}
}

// message and packet sender
func (c *Client) sender() error {
	for {
		select {
		case e := <-c.outgoing:
			if e.pkt != nil {
				// send acknowledgment
				err := c.sendAck(e.pkt)
				if err != nil {
					return err // error has already been cleaned
				}
			} else if e.msg != nil {
				// send message
				err := c.sendMessage(e.msg, e.ack)
				if err != nil {
					return err // error has already been cleaned
				}
			}
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}
	}
}

/* packet handling */

// handle an incoming Connect packet
func (c *Client) processConnect(pkt *packet.Connect) error {
	// save id
	c.id = pkt.ClientID

	// authenticate
	ok, err := c.backend.Authenticate(c, pkt.Username, pkt.Password)
	if err != nil {
		return c.die(BackendError, err)
	}

	// prepare connack packet
	connack := packet.NewConnack()
	connack.ReturnCode = packet.ConnectionAccepted
	connack.SessionPresent = false

	// check authentication
	if !ok {
		// set return code
		connack.ReturnCode = packet.NotAuthorized

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
	s, resumed, err := c.backend.Setup(c, pkt.ClientID, pkt.CleanSession)
	if err != nil {
		return c.die(BackendError, err)
	} else if s == nil {
		return c.die(BackendError, ErrMissingSession)
	}

	// set session present
	connack.SessionPresent = !pkt.CleanSession && resumed

	// assign session
	c.session = s

	// set default packet prefetch
	if c.PacketPrefetch <= 0 {
		c.PacketPrefetch = 10
	}

	// set default parallel publishes
	if c.ParallelPublishes <= 0 {
		c.ParallelPublishes = 10
	}

	// set default parallel subscribes
	if c.ParallelSubscribes <= 0 {
		c.ParallelSubscribes = 10
	}

	// set default parallel dequeues
	if c.ParallelDequeues <= 0 {
		c.ParallelDequeues = 10
	}

	// prepare publish tokens
	c.publishTokens = make(chan struct{}, c.ParallelPublishes)
	for i := 0; i < c.ParallelPublishes; i++ {
		c.publishTokens <- struct{}{}
	}

	// prepare subscribe tokens
	c.subscribeTokens = make(chan struct{}, c.ParallelSubscribes)
	for i := 0; i < c.ParallelSubscribes; i++ {
		c.subscribeTokens <- struct{}{}
	}

	// prepare dequeue tokens
	c.dequeueTokens = make(chan struct{}, c.ParallelDequeues)
	for i := 0; i < c.ParallelDequeues; i++ {
		c.dequeueTokens <- struct{}{}
	}

	// crate incoming queue
	c.incoming = make(chan packet.Generic, c.PacketPrefetch)

	// create outgoing queue
	c.outgoing = make(chan outgoing, c.ParallelPublishes+c.ParallelDequeues+c.ParallelSubscribes)

	// save will if present
	if pkt.Will != nil {
		c.will = pkt.Will
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
		publish, ok := pkt.(*packet.Publish)
		if ok {
			publish.Dup = true
		}

		// send packet
		err = c.send(pkt, true)
		if err != nil {
			return c.die(TransportError, err)
		}
	}

	// restore client
	err = c.backend.Restore(c)
	if err != nil {
		return c.die(BackendError, err)
	}

	return nil
}

// handle an incoming Generic
func (c *Client) processPacket(pkt packet.Generic) error {
	// prepare error
	var err error

	// handle individual packets
	switch typedPkt := pkt.(type) {
	case *packet.Subscribe:
		err = c.processSubscribe(typedPkt)
	case *packet.Unsubscribe:
		err = c.processUnsubscribe(typedPkt)
	case *packet.Publish:
		err = c.processPublish(typedPkt)
	case *packet.Puback:
		err = c.processPubackAndPubcomp(typedPkt.ID)
	case *packet.Pubcomp:
		err = c.processPubackAndPubcomp(typedPkt.ID)
	case *packet.Pubrec:
		err = c.processPubrec(typedPkt.ID)
	case *packet.Pubrel:
		err = c.processPubrel(typedPkt.ID)
	case *packet.Pingreq:
		err = c.processPingreq()
	case *packet.Disconnect:
		err = c.processDisconnect()
	}

	// return eventual error
	if err != nil {
		return err // error has already been cleaned
	}

	return nil
}

// handle an incoming Pingreq packet
func (c *Client) processPingreq() error {
	// send a pingresp packet
	err := c.send(packet.NewPingresp(), true)
	if err != nil {
		return c.die(TransportError, err)
	}

	return nil
}

// handle an incoming Subscribe packet
func (c *Client) processSubscribe(pkt *packet.Subscribe) error {
	// acquire subscribe token
	select {
	case <-c.subscribeTokens:
		// continue
	case <-c.tomb.Dying():
		return tomb.ErrDying
	}

	// prepare suback packet
	suback := packet.NewSuback()
	suback.ReturnCodes = make([]byte, len(pkt.Subscriptions))
	suback.ID = pkt.ID

	// set granted qos
	for i, subscription := range pkt.Subscriptions {
		suback.ReturnCodes[i] = subscription.QOS
	}

	// subscribe client to queue
	err := c.backend.Subscribe(c, pkt.Subscriptions, func() {
		select {
		case c.outgoing <- outgoing{pkt: suback}:
		case <-c.tomb.Dying():
		}
	})
	if err != nil {
		return c.die(BackendError, err)
	}

	return nil
}

// handle an incoming Unsubscribe packet
func (c *Client) processUnsubscribe(pkt *packet.Unsubscribe) error {
	// acquire subscribe token
	select {
	case <-c.subscribeTokens:
		// continue
	case <-c.tomb.Dying():
		return tomb.ErrDying
	}

	// prepare unsuback packet
	unsuback := packet.NewUnsuback()
	unsuback.ID = pkt.ID

	// unsubscribe topics
	err := c.backend.Unsubscribe(c, pkt.Topics, func() {
		select {
		case c.outgoing <- outgoing{pkt: unsuback}:
		case <-c.tomb.Dying():
		}
	})
	if err != nil {
		return c.die(BackendError, err)
	}

	return nil
}

// handle an incoming Publish packet
func (c *Client) processPublish(publish *packet.Publish) error {
	// handle qos 0 flow
	if publish.Message.QOS == 0 {
		// publish message to others
		err := c.backend.Publish(c, &publish.Message, nil)
		if err != nil {
			return c.die(BackendError, err)
		}

		c.backend.Log(MessagePublished, c, nil, &publish.Message, nil)
	}

	// handle qos 1 flow
	if publish.Message.QOS == 1 {
		// prepare puback
		puback := packet.NewPuback()
		puback.ID = publish.ID

		// acquire publish token
		select {
		case <-c.publishTokens:
			// continue
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}

		// publish message to others and queue puback if ack is called
		err := c.backend.Publish(c, &publish.Message, func() {
			c.backend.Log(MessageAcknowledged, c, nil, &publish.Message, nil)

			select {
			case c.outgoing <- outgoing{pkt: puback}:
			case <-c.tomb.Dying():
			}
		})
		if err != nil {
			return c.die(BackendError, err)
		}

		c.backend.Log(MessagePublished, c, nil, &publish.Message, nil)
	}

	// handle qos 2 flow
	if publish.Message.QOS == 2 {
		// store packet
		err := c.session.SavePacket(session.Incoming, publish)
		if err != nil {
			return c.die(SessionError, err)
		}

		// prepare pubrec packet
		pubrec := packet.NewPubrec()
		pubrec.ID = publish.ID

		// signal qos 2 pubrec
		err = c.send(pubrec, true)
		if err != nil {
			return c.die(TransportError, err)
		}
	}

	return nil
}

// handle an incoming Puback or Pubcomp packet
func (c *Client) processPubackAndPubcomp(id packet.ID) error {
	// remove packet from store
	err := c.session.DeletePacket(session.Outgoing, id)
	if err != nil {
		return c.die(SessionError, err)
	}

	// put back dequeue token
	c.dequeueTokens <- struct{}{}

	return nil
}

// handle an incoming Pubrec packet
func (c *Client) processPubrec(id packet.ID) error {
	// allocate packet
	pubrel := packet.NewPubrel()
	pubrel.ID = id

	// overwrite stored Publish with the Pubrel packet
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

// handle an incoming Pubrel packet
func (c *Client) processPubrel(id packet.ID) error {
	// get publish packet from store
	pkt, err := c.session.LookupPacket(session.Incoming, id)
	if err != nil {
		return c.die(SessionError, err)
	}

	// get packet from store
	publish, ok := pkt.(*packet.Publish)
	if !ok {
		return nil // ignore a wrongly sent Pubrel packet
	}

	// prepare pubcomp packet
	pubcomp := packet.NewPubcomp()
	pubcomp.ID = publish.ID

	// the pubrec packet will be cleared from the session once the pubcomp
	// has been sent

	// acquire publish token
	select {
	case <-c.publishTokens:
		// continue
	case <-c.tomb.Dying():
		return tomb.ErrDying
	}

	// publish message to others and queue pubcomp if ack is called
	err = c.backend.Publish(c, &publish.Message, func() {
		c.backend.Log(MessageAcknowledged, c, nil, &publish.Message, nil)

		select {
		case c.outgoing <- outgoing{pkt: pubcomp}:
		case <-c.tomb.Dying():
		}
	})
	if err != nil {
		return c.die(BackendError, err)
	}

	c.backend.Log(MessagePublished, c, nil, &publish.Message, nil)

	return nil
}

// handle an incoming Disconnect packet
func (c *Client) processDisconnect() error {
	// clear will
	c.will = nil

	// mark client as cleanly disconnected
	atomic.StoreUint32(&c.state, clientDisconnected)

	// close underlying connection (triggers cleanup)
	c.conn.Close()

	c.backend.Log(ClientDisconnected, c, nil, nil, nil)

	return ErrClientDisconnected
}

/* helpers */

// send messages
func (c *Client) sendMessage(msg *packet.Message, ack Ack) error {
	// prepare publish packet
	publish := packet.NewPublish()
	publish.Message = *msg

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

	// acknowledge message since it has been stored in the session if a quality
	// of service > 0 is requested
	if ack != nil {
		ack()

		c.backend.Log(MessageAcknowledged, c, nil, msg, nil)
	}

	// send packet
	err := c.send(publish, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	// immediately put back dequeue token for qos 0 messages
	if publish.Message.QOS == 0 {
		c.dequeueTokens <- struct{}{}
	}

	c.backend.Log(MessageForwarded, c, nil, msg, nil)

	return nil
}

// send an acknowledgment
func (c *Client) sendAck(pkt packet.Generic) error {
	// send packet
	err := c.send(pkt, true)
	if err != nil {
		return err // error already handled
	}

	// remove pubrec from session
	if pubcomp, ok := pkt.(*packet.Pubcomp); ok {
		err = c.session.DeletePacket(session.Incoming, pubcomp.ID)
		if err != nil {
			return c.die(SessionError, err)
		}
	}

	// check ack type
	switch pkt.(type) {
	case *packet.Suback, *packet.Unsuback:
		// put back subscribe token
		c.subscribeTokens <- struct{}{}
	case *packet.Puback, *packet.Pubcomp:
		// put back publish token
		c.publishTokens <- struct{}{}
	}

	return nil
}

// send a packet
func (c *Client) send(pkt packet.Generic, async bool) error {
	// send packet
	err := c.conn.Send(pkt, async)
	if err != nil {
		return err
	}

	c.backend.Log(PacketSent, c, pkt, nil, nil)

	return nil
}

/* error handling and logging */

// used for closing and cleaning up from internal goroutines
func (c *Client) die(event LogEvent, err error) error {
	// log error
	c.backend.Log(event, c, nil, nil, err)

	// close connection if requested
	c.conn.Close()

	return err
}

// will try to cleanup as many resources as possible
func (c *Client) cleanup() {
	// check if not cleanly connected and will is present
	if atomic.LoadUint32(&c.state) == clientConnected && c.will != nil {
		// publish message to others
		err := c.backend.Publish(c, c.will, nil)
		if err != nil {
			c.backend.Log(BackendError, c, nil, nil, err)
		}

		c.backend.Log(MessagePublished, c, nil, c.will, nil)
	}

	// remove client from the queue
	if atomic.LoadUint32(&c.state) >= clientConnected {
		err := c.backend.Terminate(c)
		if err != nil {
			c.backend.Log(BackendError, c, nil, nil, err)
		}
	}

	c.backend.Log(LostConnection, c, nil, nil, nil)
}
