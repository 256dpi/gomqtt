package broker

import (
	"errors"
	"sync"

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
	// connection when the call returns.
	Setup(client *Client, id string) (Session, bool, error)

	// QueueOffline is called after the clients stored subscriptions have been
	// resubscribed. It should be used to trigger a background process that
	// forwards all missed messages.
	QueueOffline(*Client) error

	// Subscribe should subscribe the passed client to the specified topic and
	// call Publish with any incoming messages.
	Subscribe(*Client, *packet.Subscription) error

	// Unsubscribe should unsubscribe the passed client from the specified topic.
	Unsubscribe(client *Client, topic string) error

	// Receive is called by the Client repeatedly to obtain the next message.
	// If the call returns no message and no error, the client will be closed
	// cleanly.
	Receive(*Client, <-chan struct{}) (*packet.Message, error)

	// StoreRetained should store the specified message.
	StoreRetained(*Client, *packet.Message) error

	// ClearRetained should remove the stored messages for the given topic.
	ClearRetained(client *Client, topic string) error

	// QueueRetained is called after acknowledging a subscription and should be
	// used to trigger a background process that forwards all retained messages.
	QueueRetained(client *Client, topic string) error

	// Publish should forward the passed message to all other clients that hold
	// a subscription that matches the messages topic. It should also add the
	// message to all sessions that have a matching offline subscription.
	Publish(*Client, *packet.Message) error

	// Terminate is called when the client goes offline. Terminate should
	// unsubscribe the passed client from all previously subscribed topics. The
	// backend may also convert a clients subscriptions to offline subscriptions.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that client as the broker will close the connection when the call
	// returns.
	Terminate(*Client) error
}

type memorySession struct {
	*session.MemorySession

	queue chan *packet.Message

	owner *Client
	kill  chan struct{}
	done  chan struct{}
}

func newMemorySession(backlog int) *memorySession {
	return &memorySession{
		MemorySession: session.NewMemorySession(),
		queue:         make(chan *packet.Message, backlog),
		kill:          make(chan struct{}, 1),
		done:          make(chan struct{}, 1),
	}
}

func (s *memorySession) reset() {
	s.kill = make(chan struct{}, 1)
	s.done = make(chan struct{}, 1)
}

// ErrQueueFull is returned to a client that attempts to write to a full queue.
var ErrQueueFull = errors.New("queue full")

// ErrKilled is returned by to a client that is killed by the broker.
var ErrKilled = errors.New("killed")

// A MemoryBackend stores everything in memory.
type MemoryBackend struct {
	// The maximal size of the session queue.
	SessionQueueSize int

	// A map of username and passwords that grant read and write access.
	Credentials      map[string]string

	storedSessions    map[string]*memorySession
	temporarySessions map[*Client]*memorySession
	retainedMessages  *topic.Tree

	globalMutex sync.Mutex
	setupMutex  sync.Mutex
}

// NewMemoryBackend returns a new MemoryBackend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		SessionQueueSize: 100,
		storedSessions:    make(map[string]*memorySession),
		temporarySessions: make(map[*Client]*memorySession),
		retainedMessages:  topic.NewTree(),
	}
}

// Authenticate authenticates a clients credentials by matching them to the
// saved Credentials map.
func (m *MemoryBackend) Authenticate(client *Client, user, password string) (bool, error) {
	// mutex locking not needed

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

	// retrieve stored session
	sess, ok := m.storedSessions[id]

	// kill existing client if existing
	if ok && sess.owner != nil {
		// send signal
		close(sess.kill)

		// release global mutex (allow publish and termination)
		m.globalMutex.Unlock()

		// wait for client to close
		<-sess.done // TODO: Timeout?

		// acquire mutex again
		m.globalMutex.Lock()

		// reload stored session
		sess, ok = m.storedSessions[id]
	}

	// TODO: Handle clean sessions.

	// reuse if (still) existing
	if ok {
		// reset session
		sess.reset()
		sess.owner = client

		return sess, true, nil
	}

	// otherwise create fresh session
	sess = newMemorySession(m.SessionQueueSize)
	sess.owner = client

	// save session
	m.storedSessions[id] = sess

	return sess, false, nil
}

// QueueOffline will begin with forwarding all missed messages in a separate
// goroutine.
func (m *MemoryBackend) QueueOffline(client *Client) error {
	// not needed as misses messages will be received in Receive()

	return nil
}

// Subscribe will subscribe the passed client to the specified topic.
func (m *MemoryBackend) Subscribe(client *Client, sub *packet.Subscription) error {
	// the subscription will be added to the session by the broker

	return nil
}

// Unsubscribe will unsubscribe the passed client from the specified topic.
func (m *MemoryBackend) Unsubscribe(client *Client, topic string) error {
	// the subscription will be removed to the session by the broker

	return nil
}

// Receive will get the next message from the queue.
func (m *MemoryBackend) Receive(client *Client, close <-chan struct{}) (*packet.Message, error) {
	// mutex locking not needed

	// get session
	sess := client.session.(*memorySession)

	// get next message from queue
	select {
	case msg := <-sess.queue:
		return msg, nil
	case <-close:
		return nil, nil
	case <-sess.kill:
		return nil, ErrKilled
	}
}

// StoreRetained will store the specified message.
func (m *MemoryBackend) StoreRetained(client *Client, msg *packet.Message) error {
	// mutex locking not needed

	// set retained message
	m.retainedMessages.Set(msg.Topic, msg.Copy())

	return nil
}

// ClearRetained will remove the stored messages for the given topic.
func (m *MemoryBackend) ClearRetained(client *Client, topic string) error {
	// mutex locking not needed

	// clear retained message
	m.retainedMessages.Empty(topic)

	return nil
}

// QueueRetained will queue all retained messages matching the given topic.
func (m *MemoryBackend) QueueRetained(client *Client, topic string) error {
	// get retained messages
	values := m.retainedMessages.Search(topic)

	// publish messages
	for _, value := range values {
		select {
		case client.session.(*memorySession).queue <- value.(*packet.Message):
		default:
			return ErrQueueFull
		}
	}

	return nil
}

// Publish will forward the passed message to all other subscribed clients. It
// will also add the message to all sessions that have a matching offline
// subscription.
func (m *MemoryBackend) Publish(client *Client, msg *packet.Message) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// add message to temporary sessions
	for _, sess := range m.temporarySessions {
		if sub, _ := sess.LookupSubscription(msg.Topic); sub != nil {
			select {
			case sess.queue <- msg:
			default:
				return ErrQueueFull
			}
		}
	}

	// add message to stored sessions
	for _, sess := range m.storedSessions {
		if sub, _ := sess.LookupSubscription(msg.Topic); sub != nil {
			select {
			case sess.queue <- msg:
			default:
				return ErrQueueFull
			}
		}
	}

	return nil
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
	sess := client.session.(*memorySession)

	// release session
	sess.owner = nil

	// delete stored session if clean is requested
	if client.CleanSession() {
		delete(m.storedSessions, client.ClientID())
	}

	// remove temporary session
	delete(m.temporarySessions, client)

	// signal exit
	close(sess.done)

	return nil
}

// Close will close the backend and make all clients go away.
func (m *MemoryBackend) Close() {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// add close temporary sessions
	for _, sess := range m.temporarySessions {
		close(sess.kill)
		<-sess.done // TODO: Timeout?
	}
}
