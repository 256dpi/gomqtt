package broker

import (
	"errors"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/topic"
)

// A Session is used to persist incoming/outgoing packets, subscriptions and the
// will.
type Session interface {
	// NextID should return the next id for outgoing packets.
	NextID() packet.ID

	// SavePacket should store a packet in the session. An eventual existing
	// packet with the same id should be quietly overwritten.
	SavePacket(session.Direction, packet.GenericPacket) error

	// LookupPacket should retrieve a packet from the session using the packet id.
	LookupPacket(session.Direction, packet.ID) (packet.GenericPacket, error)

	// DeletePacket should remove a packet from the session. The method should
	// not return an error if no packet with the specified id does exists.
	DeletePacket(session.Direction, packet.ID) error

	// AllPackets should return all packets currently saved in the session. This
	// method is used to resend stored packets when the session is resumed.
	AllPackets(session.Direction) ([]packet.GenericPacket, error)

	// SaveSubscription should store the subscription in the session. An eventual
	// subscription with the same topic should be quietly overwritten.
	SaveSubscription(*packet.Subscription) error

	// LookupSubscription should match a topic against the stored subscriptions
	// and eventually return the first found subscription.
	LookupSubscription(topic string) (*packet.Subscription, error)

	// DeleteSubscription should remove the subscription from the session. The
	// method should not return an error if no subscription with the specified
	// topic does exist.
	DeleteSubscription(topic string) error

	// AllSubscriptions should return all subscriptions currently saved in the
	// session. This method is used to restore a clients subscriptions when the
	// session is resumed.
	AllSubscriptions() ([]*packet.Subscription, error)

	// SaveWill should store the will message.
	SaveWill(*packet.Message) error

	// LookupWill should retrieve the will message.
	LookupWill() (*packet.Message, error)

	// ClearWill should remove the will message from the store.
	ClearWill() error

	// Reset should completely reset the session.
	Reset() error
}

// Ack is executed by the Backend or Client to signal that a message will be
// delivered under the selected qos level and is therefore safe to be deleted
// from either queue.
type Ack func(message *packet.Message)

// A Backend provides the effective brokering functionality to its clients.
type Backend interface {
	// Authenticate should authenticate the client using the user and password
	// values and return true if the client is eligible to continue or false
	// when the broker should terminate the connection.
	Authenticate(client *Client, user, password string) (bool, error)

	// Setup is called when a new client comes online and is successfully
	// authenticated. Setup should return the already stored session for the
	// supplied id or create and return a new one. If the supplied id has a zero
	// length, a new temporary session should returned that is not stored
	// further. The backend may also close any existing clients that use the
	// same client id.
	//
	// Note: In this call the Backend may also allocate other resources and
	// setup the client for further usage as the broker will acknowledge the
	// connection when the call returns. The Terminate function is called for
	// every client that Setup has been called for.
	Setup(client *Client, id string) (Session, bool, error)

	// Restored is called after the client has restored packets and
	// subscriptions from the session. The client will begin with processing
	// incoming packets and queued messages.
	Restored(client *Client) error

	// Subscribe should subscribe the passed client to the specified topic and
	// call Publish with any incoming messages. The subscription will also be
	// added to the session if the call returns without an error.
	//
	// Retained messages that match the supplied subscription should be added to
	// a temporary queue that is also drained when Dequeue is called. Retained
	// messages are not part of the stored session queue as they are anyway
	// redelivered using the stored subscription mechanism.
	//
	// Subscribe is also called to resubscribe stored subscriptions between calls
	// to Setup and Restored. Retained messages that are delivered as a result of
	// resubscribing a stored subscription must be delivered with the retain flag
	// set to false.
	Subscribe(client *Client, sub *packet.Subscription, stored bool) error

	// Unsubscribe should unsubscribe the passed client from the specified topic.
	// The subscription will also be removed from the session if the call returns
	// without an error.
	Unsubscribe(client *Client, topic string) error

	// Publish should forward the passed message to all other clients that hold
	// a subscription that matches the messages topic. It should also add the
	// message to all sessions that have a matching offline subscription. The
	// later may only apply to messages with a QoS greater than 0.
	//
	// If the retained flag is set, messages with a payload should replace the
	// currently retained message. Otherwise, the currently retained message
	// should be removed. The flag should be cleared before publishing the
	// message to subscribed clients.
	//
	// Note: If the backend does not return an error the message will be
	// immediately acknowledged by the client and removed from the session.
	Publish(client *Client, msg *packet.Message) error

	// TODO: Publish with ack.

	// Dequeue is called by the Client repeatedly to obtain the next message.
	// The backend must return no message and no error if the supplied channel
	// is closed. The returned Ack is executed by the Backend to signal that the
	// message is being delivered under the selected qos level and is therefore
	// safe to be deleted from the queue.
	Dequeue(client *Client, close <-chan struct{}) (*packet.Message, Ack, error)

	// Terminate is called when the client goes offline. Terminate should
	// unsubscribe the passed client from all previously subscribed topics. The
	// backend may also convert a clients subscriptions to offline subscriptions.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that client as the broker will close the connection when the call
	// returns.
	Terminate(client *Client) error
}

type memorySession struct {
	*session.MemorySession

	stored    chan *packet.Message
	temporary chan *packet.Message

	owner *Client
	kill  chan struct{}
	done  chan struct{}
}

func newMemorySession(backlog int) *memorySession {
	return &memorySession{
		MemorySession: session.NewMemorySession(),
		stored:        make(chan *packet.Message, backlog),
		temporary:     make(chan *packet.Message, backlog),
		kill:          make(chan struct{}, 1),
		done:          make(chan struct{}, 1),
	}
}

func (s *memorySession) reuse() {
	s.temporary = make(chan *packet.Message, cap(s.temporary))
	s.kill = make(chan struct{}, 1)
	s.done = make(chan struct{}, 1)
}

// ErrQueueFull is returned to a client that attempts two write to its own full
// queue, which would result in a deadlock.
var ErrQueueFull = errors.New("queue full")

// ErrKilled is returned to a client that is killed by the broker.
var ErrKilled = errors.New("killed")

// ErrClosing is returned to a client if the backend is closing.
var ErrClosing = errors.New("closing")

// ErrKillTimeout is returned to a client if the killed existing client does not
// close in time.
var ErrKillTimeout = errors.New("kill timeout")

// A MemoryBackend stores everything in memory.
type MemoryBackend struct {
	// The maximal size of the session queue.
	SessionQueueSize int

	// The time after an error is returned while waiting on an killed existing
	// client to exit.
	KillTimeout time.Duration

	// A map of username and passwords that grant read and write access.
	Credentials map[string]string

	activeClients     map[string]*Client
	storedSessions    map[string]*memorySession
	temporarySessions map[*Client]*memorySession
	retainedMessages  *topic.Tree

	globalMutex sync.Mutex
	setupMutex  sync.Mutex
	closing     bool
}

// NewMemoryBackend returns a new MemoryBackend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		SessionQueueSize:  100,
		KillTimeout:       5 * time.Second,
		activeClients:     make(map[string]*Client),
		storedSessions:    make(map[string]*memorySession),
		temporarySessions: make(map[*Client]*memorySession),
		retainedMessages:  topic.NewTree(),
	}
}

// Authenticate authenticates a clients credentials by matching them to the
// saved Credentials map.
func (m *MemoryBackend) Authenticate(client *Client, user, password string) (bool, error) {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// return error if closing
	if m.closing {
		return false, ErrClosing
	}

	// allow all if there are no credentials
	if m.Credentials == nil {
		return true, nil
	}

	// check login
	if pw, ok := m.Credentials[user]; ok && pw == password {
		return true, nil
	}

	return false, nil
}

// Setup returns the already stored session for the supplied id or creates and
// returns a new one. If the supplied id has a zero length, a new session is
// returned that is not stored further. Furthermore, it will disconnect any client
// connected with the same client id.
func (m *MemoryBackend) Setup(client *Client, id string) (Session, bool, error) {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// acquire setup mutex
	m.setupMutex.Lock()
	defer m.setupMutex.Unlock()

	// return error if closing
	if m.closing {
		return nil, false, ErrClosing
	}

	// return a new temporary session if id is zero
	if len(id) == 0 {
		// create session
		sess := newMemorySession(m.SessionQueueSize)
		sess.owner = client

		// save session
		m.temporarySessions[client] = sess

		return sess, false, nil
	}

	// client id is available

	// retrieve existing client
	existingSession, ok := m.storedSessions[id]
	if !ok {
		if existingClient, ok2 := m.activeClients[id]; ok2 {
			existingSession, ok = m.temporarySessions[existingClient]
		}
	}

	// kill existing client if session is taken
	if ok && existingSession.owner != nil {
		// send signal
		close(existingSession.kill)

		// release global mutex to allow publish and termination, but leave the
		// setup mutex to prevent setups
		m.globalMutex.Unlock()

		// wait for client to close
		select {
		case <-existingSession.done:
			// continue
		case <-time.After(m.KillTimeout):
			return nil, false, ErrKillTimeout
		}

		// acquire mutex again
		m.globalMutex.Lock()
	}

	// delete any stored session and return temporary if requested
	if client.CleanSession() {
		// delete any stored session
		delete(m.storedSessions, id)

		// create new session
		sess := newMemorySession(m.SessionQueueSize)
		sess.owner = client

		// save session
		m.temporarySessions[client] = sess

		// save client
		m.activeClients[id] = client

		return sess, false, nil
	}

	// attempt to reuse a stored session
	storedSession, ok := m.storedSessions[id]
	if ok {
		// reuse session
		storedSession.reuse()
		storedSession.owner = client

		// save client
		m.activeClients[id] = client

		return storedSession, true, nil
	}

	// otherwise create fresh session
	storedSession = newMemorySession(m.SessionQueueSize)
	storedSession.owner = client

	// save session
	m.storedSessions[id] = storedSession

	// save client
	m.activeClients[id] = client

	return storedSession, false, nil
}

// Restored is not needed at the moment.
func (m *MemoryBackend) Restored(client *Client) error {
	return nil
}

// Subscribe will queue retained messages for the passed subscription.
func (m *MemoryBackend) Subscribe(client *Client, sub *packet.Subscription, stored bool) error {
	// mutex locking not needed

	// get retained messages
	values := m.retainedMessages.Search(sub.Topic)

	// publish messages
	for _, value := range values {
		// get message
		msg := value.(*packet.Message)

		// unset retained flag for stored subscriptions
		if stored {
			msg = msg.Copy()
			msg.Retain = false
		}

		select {
		case client.Session().(*memorySession).temporary <- msg:
		default:
			return ErrQueueFull
		}
	}

	return nil
}

// Unsubscribe is not needed as the subscription will be removed from the session
// by the broker
func (m *MemoryBackend) Unsubscribe(client *Client, topic string) error {
	return nil
}

// Publish will forward the passed message to all other subscribed clients. It
// will also add the message to all sessions that have a matching offline
// subscription.
func (m *MemoryBackend) Publish(client *Client, msg *packet.Message) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// check retain flag
	if msg.Retain {
		if len(msg.Payload) > 0 {
			// retain message
			m.retainedMessages.Set(msg.Topic, msg.Copy())
		} else {
			// clear already retained message
			m.retainedMessages.Empty(msg.Topic)
		}
	}

	// define default queue accessor
	queue := func(s *memorySession) chan *packet.Message {
		return s.temporary
	}

	// check qos flag
	if msg.QOS > 0 {
		queue = func(s *memorySession) chan *packet.Message {
			return s.stored
		}
	}

	// check retained flag
	if msg.Retain {
		// redefine queue accessor
		queue = func(s *memorySession) chan *packet.Message {
			return s.temporary
		}

		// reset flag
		msg.Retain = false
	}

	// add message to temporary sessions
	for _, sess := range m.temporarySessions {
		if sub, _ := sess.LookupSubscription(msg.Topic); sub != nil {
			// detect deadlock when adding to own queue
			if sess.owner == client {
				select {
				case queue(sess) <- msg:
				default:
					return ErrQueueFull
				}
			} else {
				queue(sess) <- msg
			}
		}
	}

	// add message to stored sessions
	for _, sess := range m.storedSessions {
		if sub, _ := sess.LookupSubscription(msg.Topic); sub != nil {
			// detect deadlock when adding to own queue
			if sess.owner == client {
				select {
				case queue(sess) <- msg:
				default:
					return ErrQueueFull
				}
			} else {
				queue(sess) <- msg
			}
		}
	}

	return nil
}

// Dequeue will get the next message from the queue.
func (m *MemoryBackend) Dequeue(client *Client, close <-chan struct{}) (*packet.Message, Ack, error) {
	// mutex locking not needed

	// get session
	sess := client.Session().(*memorySession)

	// TODO: Add ack support.

	// get next message from queue
	select {
	case msg := <-sess.temporary:
		return msg, nil, nil
	case msg := <-sess.stored:
		return msg, nil, nil
	case <-close:
		return nil, nil, nil
	case <-sess.kill:
		return nil, nil, ErrKilled
	}
}

// Terminate will unsubscribe the passed client from all previously subscribed
// topics. If the client connect with clean=true it will also clean the session.
// Otherwise it will create offline subscriptions for all QOS 1 and QOS 2
// subscriptions.
func (m *MemoryBackend) Terminate(client *Client) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// get session
	sess := client.Session().(*memorySession)

	// release session
	sess.owner = nil

	// remove any temporary session
	delete(m.temporarySessions, client)

	// remove any saved client
	delete(m.activeClients, client.ClientID())

	// signal exit
	close(sess.done)

	return nil
}

// Close will close all active clients and close the backend. The return value
// denotes if the timeout has been reached.
func (m *MemoryBackend) Close(timeout time.Duration) bool {
	// acquire global mutex
	m.globalMutex.Lock()

	// set closing
	m.closing = true

	// prepare channel list
	var list []chan struct{}

	// close temporary sessions
	for _, sess := range m.temporarySessions {
		close(sess.kill)
		list = append(list, sess.done)
	}

	// closed owned stored sessions
	for _, sess := range m.storedSessions {
		if sess.owner != nil {
			close(sess.kill)
			list = append(list, sess.done)
		}
	}

	// release mutex to allow termination
	m.globalMutex.Unlock()

	// return early if empty
	if len(list) == 0 {
		return true
	}

	// prepare timeout
	tm := time.After(timeout)

	// wait for clients to close
	for _, ch := range list {
		select {
		case <-ch:
			continue
		case <-tm:
			return false
		}
	}

	return true
}
