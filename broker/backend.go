package broker

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/tools"
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

	// QueueOffline should enqueue the specified offline message.
	QueueOffline(*packet.Message)

	// DequeueOffline should dequeue the next offline message.
	DequeueOffline() *packet.Message

	// Reset should completely reset the session.
	Reset() error
}

// A Backend provides effective queuing functionality to a Broker and its Clients.
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
	QueueOffline(client *Client) error

	// Subscribe should subscribe the passed client to the specified topic and
	// call Publish with any incoming messages.
	Subscribe(client *Client, topic string) error

	// Unsubscribe should unsubscribe the passed client from the specified topic.
	Unsubscribe(client *Client, topic string) error

	// StoreRetained should store the specified message.
	StoreRetained(client *Client, msg *packet.Message) error

	// ClearRetained should remove the stored messages for the given topic.
	ClearRetained(client *Client, topic string) error

	// QueueRetained is called after acknowledging a subscription and should be
	// used to trigger a background process that forwards all retained messages.
	QueueRetained(client *Client, topic string) error

	// Publish should forward the passed message to all other clients that hold
	// a subscription that matches the messages topic. It should also add the
	// message to all sessions that have a matching offline subscription.
	Publish(client *Client, msg *packet.Message) error

	// Terminate is called when the client goes offline. Terminate should
	// unsubscribe the passed client from all previously subscribed topics. The
	// backend may also convert a clients subscriptions to offline subscriptions.
	//
	// Note: The Backend may also cleanup previously allocated resources for
	// that client as the broker will close the connection when the call
	// returns.
	Terminate(client *Client) error
}

// A MemoryBackend stores everything in memory.
type MemoryBackend struct {
	Credentials map[string]string

	queue        *tools.Tree
	retained     *tools.Tree
	offlineQueue *tools.Tree

	sessions      map[string]Session
	sessionsMutex sync.Mutex

	clients map[string]*Client
}

// NewMemoryBackend returns a new MemoryBackend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		queue:        tools.NewTree(),
		retained:     tools.NewTree(),
		offlineQueue: tools.NewTree(),
		sessions:     make(map[string]Session),
		clients:      make(map[string]*Client),
	}
}

// Authenticate authenticates a clients credentials by matching them to the
// saved Credentials map.
func (m *MemoryBackend) Authenticate(client *Client, user, password string) (bool, error) {
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
	m.sessionsMutex.Lock()
	defer m.sessionsMutex.Unlock()

	// return a new temporary session if id is zero
	if len(id) == 0 {
		return session.NewMemorySession(), false, nil
	}

	// close existing client
	existingClient, ok := m.clients[id]
	if ok {
		existingClient.Close(true)
	}

	// add new client
	m.clients[id] = client

	// retrieve stored session
	s, ok := m.sessions[id]

	// when found
	if ok {
		// remove session if clean is true
		if client.CleanSession() {
			delete(m.sessions, id)
		}

		// clear offline subscriptions
		m.offlineQueue.Clear(s)

		return s, true, nil
	}

	// create fresh session
	s = session.NewMemorySession()

	// save session if not clean
	if !client.CleanSession() {
		m.sessions[id] = s
	}

	return s, false, nil
}

// QueueOffline will begin with forwarding all missed messages in a separate
// goroutine.
func (m *MemoryBackend) QueueOffline(client *Client) error {
	// send all missed messages in another goroutine
	go func() {
		for {
			// get next missed message
			msg := client.session.DequeueOffline()
			if msg == nil {
				return
			}

			// publish or add back to queue
			if !client.Publish(msg) {
				client.session.QueueOffline(msg)
				return
			}
		}
	}()

	return nil
}

// Subscribe will subscribe the passed client to the specified topic and
// begin to forward messages by calling the clients Publish method.
func (m *MemoryBackend) Subscribe(client *Client, topic string) error {
	// add subscription
	m.queue.Add(topic, client)

	return nil
}

// Unsubscribe will unsubscribe the passed client from the specified topic.
func (m *MemoryBackend) Unsubscribe(client *Client, topic string) error {
	// remove subscription
	m.queue.Remove(topic, client)

	return nil
}

// StoreRetained will store the specified message.
func (m *MemoryBackend) StoreRetained(client *Client, msg *packet.Message) error {
	// set retained message
	m.retained.Set(msg.Topic, msg.Copy())

	return nil
}

// ClearRetained will remove the stored messages for the given topic.
func (m *MemoryBackend) ClearRetained(client *Client, topic string) error {
	// clear retained message
	m.retained.Empty(topic)

	return nil
}

// QueueRetained will queue all retained messages matching the given topic.
func (m *MemoryBackend) QueueRetained(client *Client, topic string) error {
	// get retained messages
	values := m.retained.Search(topic)

	// publish messages
	for _, value := range values {
		client.Publish(value.(*packet.Message))
	}

	return nil
}

// Publish will forward the passed message to all other subscribed clients. It
// will also add the message to all sessions that have a matching offline
// subscription.
func (m *MemoryBackend) Publish(client *Client, msg *packet.Message) error {
	// publish directly to clients
	for _, v := range m.queue.Match(msg.Topic) {
		v.(*Client).Publish(msg)
	}

	// queue for offline clients
	for _, v := range m.offlineQueue.Match(msg.Topic) {
		v.(Session).QueueOffline(msg)
	}

	return nil
}

// Terminate will unsubscribe the passed client from all previously subscribed
// topics. If the client connect with clean=true it will also clean the session.
// Otherwise it will create offline subscriptions for all QOS 1 and QOS 2
// subscriptions.
func (m *MemoryBackend) Terminate(client *Client) error {
	m.sessionsMutex.Lock()
	defer m.sessionsMutex.Unlock()

	// clear all subscriptions
	m.queue.Clear(client)

	// remove client from list
	if len(client.ClientID()) > 0 {
		delete(m.clients, client.ClientID())
	}

	// return if the client requested a clean session
	if client.CleanSession() {
		return nil
	}

	// otherwise get stored subscriptions
	subscriptions, err := client.Session().AllSubscriptions()
	if err != nil {
		return err
	}

	// iterate through stored subscriptions
	for _, sub := range subscriptions {
		if sub.QOS >= 1 {
			// add offline subscription
			m.offlineQueue.Add(sub.Topic, client.Session())
		}
	}

	return nil
}
