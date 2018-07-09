package broker

import (
	"errors"
	"fmt"
	"net"
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

// ErrClosed is returned if a client is being closed by the broker.
var ErrClosed = errors.New("closed")

type outgoing struct {
	pkt packet.GenericPacket
	msg *packet.Message
	ack Ack
}

// A Client represents a remote client that is connected to the broker.
type Client struct {
	// Ref can be used to store a custom reference to an object. This is usually
	// used to attach a state object to client that is created in the Backend.
	Ref interface{}

	// PacketPrefetch may be set during Setup to control the number of packets
	// that are read by Client and made available for processing. Will default
	// to 10 if not set.
	PacketPrefetch int

	// ParallelPublishes may be set during Setup to control the number of
	// parallel calls to Publish a client can perform. Will default to 10.
	ParallelPublishes int

	// ParallelDequeues may be set during Setup to control the number of
	// parallel calls to Dequeue a client can perform. Will default to 10.
	ParallelDequeues int

	backend Backend
	logger  Logger
	conn    transport.Conn

	state        uint32
	clientID     string
	cleanSession bool
	session      Session

	incoming chan packet.GenericPacket
	outgoing chan outgoing

	publishTokens chan struct{}
	dequeueTokens chan struct{}

	tomb tomb.Tomb
	done chan struct{}
}

// NewClient takes over a connection and returns a Client.
func NewClient(backend Backend, logger Logger, conn transport.Conn) *Client {
	// create client
	c := &Client{
		state:   clientConnecting,
		backend: backend,
		logger:  logger,
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

// Close will immediately close the client.
func (c *Client) Close() {
	c.tomb.Kill(ErrClosed)
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
	c.log(NewConnection, c, nil, nil, nil)

	// prepare packet
	var pkt packet.GenericPacket

	// get first packet from connection
	pkt, err := c.conn.Receive()
	if err != nil {
		return c.die(TransportError, err)
	}

	// get connect
	connect, ok := pkt.(*packet.ConnectPacket)
	if !ok {
		return c.die(ClientError, ErrExpectedConnect)
	}

	// process connect
	err = c.processConnect(connect)
	if err != nil {
		return err // error has already been cleaned
	}

	// start receiver, dequeuer and sender
	c.tomb.Go(c.receiver)
	c.tomb.Go(c.dequeuer)
	c.tomb.Go(c.sender)

	// TODO: Possibly we could inline the receiver?

	for {
		select {
		case pkt = <-c.incoming:
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
		// receive next packet
		pkt, err := c.conn.Receive()
		if err != nil {
			return c.die(TransportError, err)
		}

		// send packet
		select {
		case c.incoming <- pkt:
			// continue
		case <-c.tomb.Dying():
			return tomb.ErrDying
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
		msg, ack, err := c.backend.Dequeue(c, c.tomb.Dying())
		if err != nil {
			return c.die(BackendError, err)
		} else if msg == nil {
			return tomb.ErrDying
		}

		// queue message
		select {
		case c.outgoing <- outgoing{msg: msg, ack: ack}:
			// continue
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}
	}
}

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

	// set session present
	connack.SessionPresent = !pkt.CleanSession && resumed

	// assign session
	c.session = s

	// set default packet prefetch
	if c.PacketPrefetch <= 0 {
		c.PacketPrefetch = 10
	}

	// set default unacked publishes
	if c.ParallelPublishes <= 0 {
		c.ParallelPublishes = 10
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

	// prepare dequeue tokens
	c.dequeueTokens = make(chan struct{}, c.ParallelDequeues)
	for i := 0; i < c.ParallelDequeues; i++ {
		c.dequeueTokens <- struct{}{}
	}

	// crate incoming queue
	c.incoming = make(chan packet.GenericPacket, c.PacketPrefetch)

	// create outgoing queue
	c.outgoing = make(chan outgoing, c.ParallelPublishes+c.ParallelDequeues)

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
	if !pkt.CleanSession { // TODO: Do anyway.
		// get stored subscriptions
		subs, err := s.AllSubscriptions()
		if err != nil {
			return c.die(SessionError, err)
		}

		// resubscribe subscriptions
		for _, sub := range subs {
			err = c.backend.Subscribe(c, sub, true)
			if err != nil {
				return c.die(BackendError, err)
			}
		}

		// begin with queueing offline messages
		err = c.backend.Restored(c)
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
		err = c.backend.Subscribe(c, &subscription, false)
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
	// handle qos 0 flow
	if publish.Message.QOS == 0 {
		// publish message to others
		err := c.backend.Publish(c, &publish.Message, nil)
		if err != nil {
			return c.die(BackendError, err)
		}

		c.log(MessagePublished, c, nil, &publish.Message, nil)
	}

	// TODO: Client might deadlock if there are too many publishes and outgoing
	// packets have not yet been sent out.

	// handle qos 1 flow
	if publish.Message.QOS == 1 {
		// prepare puback
		puback := packet.NewPubackPacket()
		puback.ID = publish.ID

		// acquire publish  token
		select {
		case <-c.publishTokens:
			// continue
		case <-c.tomb.Dying():
			return tomb.ErrDying
		}

		// publish message to others and queue puback if ack is called
		err := c.backend.Publish(c, &publish.Message, func(msg *packet.Message) {
			select {
			case c.outgoing <- outgoing{pkt: puback}:
			case <-c.tomb.Dying():
			default:
				panic(fmt.Sprintf("o: %d, p: %d, d: %d", len(c.outgoing), len(c.dequeueTokens), len(c.publishTokens)))
			}
		})
		if err != nil {
			return c.die(BackendError, err)
		}

		c.log(MessagePublished, c, nil, &publish.Message, nil)
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

		// signal qos 2 pubrec
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
	err := c.session.DeletePacket(session.Outgoing, id)
	if err != nil {
		return c.die(SessionError, err)
	}

	// put back dequeue token
	c.dequeueTokens <- struct{}{}

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
	// get publish packet from store
	pkt, err := c.session.LookupPacket(session.Incoming, id)
	if err != nil {
		return c.die(SessionError, err)
	}

	// get packet from store
	publish, ok := pkt.(*packet.PublishPacket)
	if !ok {
		return nil // ignore a wrongly sent PubrelPacket
	}

	// prepare pubcomp packet
	pubcomp := packet.NewPubcompPacket()
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
	err = c.backend.Publish(c, &publish.Message, func(msg *packet.Message) {
		select {
		case c.outgoing <- outgoing{pkt: pubcomp}:
		case <-c.tomb.Dying():
		}
	})
	if err != nil {
		return c.die(BackendError, err)
	}

	c.log(MessagePublished, c, nil, &publish.Message, nil)

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

// send messages
func (c *Client) sendMessage(msg *packet.Message, ack Ack) error {
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

	// acknowledge message since it has been stored in the session if a quality
	// of service > 0 is requested
	if ack != nil {
		ack(msg)

		c.log(MessageAcknowledged, c, nil, msg, nil)
	}

	// send packet
	err = c.send(publish, true)
	if err != nil {
		return c.die(TransportError, err)
	}

	// immediately put back dequeue token for qos 0 messages
	if publish.Message.QOS == 0 {
		c.dequeueTokens <- struct{}{}
	}

	c.log(MessageForwarded, c, nil, msg, nil)

	return nil
}

// send an acknowledgment
func (c *Client) sendAck(pkt packet.GenericPacket) error {
	// send packet
	err := c.send(pkt, true)
	if err != nil {
		return err // error already handled
	}

	// remove pubrec from session
	if pubcomp, ok := pkt.(*packet.PubcompPacket); ok {
		err = c.session.DeletePacket(session.Incoming, pubcomp.ID)
		if err != nil {
			return c.die(SessionError, err)
		}
	}

	// put back publish token
	c.publishTokens <- struct{}{}

	return nil
}

// send a packet
func (c *Client) send(pkt packet.GenericPacket, buffered bool) error {
	// send packet
	var err error
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
			// publish message to others
			err := c.backend.Publish(c, will, nil)
			if err != nil {
				c.log(BackendError, c, nil, nil, err)
			}

			c.log(MessagePublished, c, nil, will, nil)
		}
	}

	// remove client from the queue
	if atomic.LoadUint32(&c.state) >= clientConnected {
		err := c.backend.Terminate(c)
		if err != nil {
			c.log(BackendError, c, nil, nil, err)
		}
	}

	c.log(LostConnection, c, nil, nil, nil)
}
