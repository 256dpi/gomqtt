package broker

import (
	"errors"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/topic"
)

// TODO: Improve thread-safety of memorySession and Backend.

type memorySession struct {
	*session.MemorySession

	stored    chan *packet.Message
	temporary chan *packet.Message

	owner *Client
}

func newMemorySession(backlog int) *memorySession {
	return &memorySession{
		MemorySession: session.NewMemorySession(),
		stored:        make(chan *packet.Message, backlog),
		temporary:     make(chan *packet.Message, backlog),
	}
}

func (s *memorySession) reuse() {
	s.temporary = make(chan *packet.Message, cap(s.temporary))
}

// ErrQueueFull is returned to a client that attempts two write to its own full
// queue, which would result in a deadlock.
var ErrQueueFull = errors.New("queue full")

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
func (m *MemoryBackend) Setup(client *Client, id string, clean bool) (Session, bool, error) {
	// acquire setup mutex
	m.setupMutex.Lock()
	defer m.setupMutex.Unlock()

	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

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
		// close client
		existingSession.owner.Close()

		// release global mutex to allow publish and termination, but leave the
		// setup mutex to prevent setups
		m.globalMutex.Unlock()

		// wait for client to close
		var err error
		select {
		case <-existingSession.owner.Closed():
			// continue
		case <-time.After(m.KillTimeout):
			err = ErrKillTimeout
		}

		// acquire mutex again
		m.globalMutex.Lock()

		// return err if set
		if err != nil {
			return nil, false, err
		}
	}

	// delete any stored session and return a temporary session if a clean
	// session is requested
	if clean {
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
func (m *MemoryBackend) Publish(client *Client, msg *packet.Message, ack Ack) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// TODO: Remove mutex as it deadlocks if a Publish is waiting on a client.

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
			if sess.owner == client {
				// detect deadlock when adding to own queue
				select {
				case queue(sess) <- msg:
				default:
					return ErrQueueFull
				}
			} else {
				// wait for room since client is online
				select {
				case queue(sess) <- msg:
				case <-sess.owner.Closed():
				case <-client.Closed():
				}
			}
		}
	}

	// add message to stored sessions
	for _, sess := range m.storedSessions {
		if sub, _ := sess.LookupSubscription(msg.Topic); sub != nil {
			if sess.owner == client {
				// detect deadlock when adding to own queue
				select {
				case queue(sess) <- msg:
				default:
					return ErrQueueFull
				}
			} else if sess.owner != nil {
				// wait for room if client is online
				select {
				case queue(sess) <- msg:
				case <-sess.owner.Closed():
				case <-client.Closed():
				}
			} else {
				// ignore message if queue is full
				select {
				case queue(sess) <- msg:
				default:
				}
			}
		}
	}

	// call ack if available
	if ack != nil {
		ack(msg)
	}

	return nil
}

// Dequeue will get the next message from the queue.
func (m *MemoryBackend) Dequeue(client *Client) (*packet.Message, Ack, error) {
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
	case <-client.Closing():
		return nil, nil, nil
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

	// release session if available
	sess, ok := client.Session().(*memorySession)
	if ok && sess != nil {
		sess.owner = nil
	}

	// remove any temporary session
	delete(m.temporarySessions, client)

	// remove any saved client
	delete(m.activeClients, client.ID())

	return nil
}

// Close will close all active clients and close the backend. The return value
// denotes if the timeout has been reached.
func (m *MemoryBackend) Close(timeout time.Duration) bool {
	// acquire global mutex
	m.globalMutex.Lock()

	// set closing
	m.closing = true

	// prepare list
	var clients []*Client

	// close temporary sessions
	for _, sess := range m.temporarySessions {
		sess.owner.Close()
		clients = append(clients, sess.owner)
	}

	// closed owned stored sessions
	for _, sess := range m.storedSessions {
		if sess.owner != nil {
			sess.owner.Close()
			clients = append(clients, sess.owner)
		}
	}

	// release mutex to allow termination
	m.globalMutex.Unlock()

	// return early if empty
	if len(clients) == 0 {
		return true
	}

	// prepare timeout
	tm := time.After(timeout)

	// wait for clients to close
	for _, client := range clients {
		select {
		case <-client.Closed():
			continue
		case <-tm:
			return false
		}
	}

	return true
}
