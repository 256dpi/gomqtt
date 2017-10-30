package broker

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/tools"
)

// Direction denotes a packets direction.
type Direction int

const (
	// Outgoing packets are being be sent.
	Outgoing Direction = iota

	// Incoming packets are being received.
	Incoming
)

// A Session is used to persist incoming/outgoing packets, subscriptions and the
// will.
type Session interface {
	// NextID should return the next id for outgoing packets.
	NextID() packet.ID

	// SavePacket should store a packet in the session. An eventual existing
	// packet with the same id should be quietly overwritten.
	SavePacket(Direction, packet.GenericPacket) error

	// LookupPacket should retrieve a packet from the session using the packet id.
	LookupPacket(Direction, packet.ID) (packet.GenericPacket, error)

	// DeletePacket should remove a packet from the session. The method should
	// not return an error if no packet with the specified id does exists.
	DeletePacket(Direction, packet.ID) error

	// AllPackets should return all packets currently saved in the session. This
	// method is used to resend stored packets when the session is resumed.
	AllPackets(Direction) ([]packet.GenericPacket, error)

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
	SaveWill(msg *packet.Message) error

	// LookupWill should retrieve the will message.
	LookupWill() (*packet.Message, error)

	// ClearWill should remove the will message from the store.
	ClearWill() error

	// Reset should completely reset the session.
	Reset() error
}

// A MemorySession stores packets, subscriptions and the will in memory.
type MemorySession struct {
	counter       *tools.Counter
	incStore      *tools.Store
	outStore      *tools.Store
	subscriptions *tools.Tree
	offlineStore  *tools.Queue

	will      *packet.Message
	willMutex sync.Mutex
}

// NewMemorySession returns a new MemorySession.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		counter:       tools.NewCounter(),
		incStore:      tools.NewStore(),
		outStore:      tools.NewStore(),
		subscriptions: tools.NewTree(),
		offlineStore:  tools.NewQueue(100),
	}
}

// NextID will return the next id for outgoing packets.
func (s *MemorySession) NextID() packet.ID {
	return s.counter.Next()
}

// SavePacket will store a packet in the session. An eventual existing
// packet with the same id gets quietly overwritten.
func (s *MemorySession) SavePacket(dir Direction, pkt packet.GenericPacket) error {
	s.storeForDirection(dir).Save(pkt)
	return nil
}

// LookupPacket will retrieve a packet from the session using a packet id.
func (s *MemorySession) LookupPacket(dir Direction, id packet.ID) (packet.GenericPacket, error) {
	return s.storeForDirection(dir).Lookup(id), nil
}

// DeletePacket will remove a packet from the session. The method will not
// return an error if no packet with the specified id exists.
func (s *MemorySession) DeletePacket(dir Direction, id packet.ID) error {
	s.storeForDirection(dir).Delete(id)
	return nil
}

// AllPackets will return all packets currently saved in the session.
func (s *MemorySession) AllPackets(dir Direction) ([]packet.GenericPacket, error) {
	return s.storeForDirection(dir).All(), nil
}

// SaveSubscription will store the subscription in the session. An eventual
// subscription with the same topic gets quietly overwritten.
func (s *MemorySession) SaveSubscription(sub *packet.Subscription) error {
	s.subscriptions.Set(sub.Topic, sub)
	return nil
}

// LookupSubscription will match a topic against the stored subscriptions and
// eventually return the first found subscription.
func (s *MemorySession) LookupSubscription(topic string) (*packet.Subscription, error) {
	values := s.subscriptions.Match(topic)

	if len(values) > 0 {
		return values[0].(*packet.Subscription), nil
	}

	return nil, nil
}

// DeleteSubscription will remove the subscription from the session. The
// method will not return an error if no subscription with the specified
// topic does exist.
func (s *MemorySession) DeleteSubscription(topic string) error {
	s.subscriptions.Empty(topic)
	return nil
}

// AllSubscriptions will return all subscriptions currently saved in the session.
func (s *MemorySession) AllSubscriptions() ([]*packet.Subscription, error) {
	var all []*packet.Subscription

	for _, value := range s.subscriptions.All() {
		all = append(all, value.(*packet.Subscription))
	}

	return all, nil
}

// SaveWill will store the will message.
func (s *MemorySession) SaveWill(newWill *packet.Message) error {
	s.willMutex.Lock()
	defer s.willMutex.Unlock()

	s.will = newWill

	return nil
}

// LookupWill will retrieve the will message.
func (s *MemorySession) LookupWill() (*packet.Message, error) {
	s.willMutex.Lock()
	defer s.willMutex.Unlock()

	return s.will, nil
}

// ClearWill will remove the will message from the store.
func (s *MemorySession) ClearWill() error {
	s.willMutex.Lock()
	defer s.willMutex.Unlock()

	s.will = nil

	return nil
}

// Reset will completely reset the session.
func (s *MemorySession) Reset() error {
	s.counter.Reset()
	s.incStore.Reset()
	s.outStore.Reset()
	s.subscriptions.Reset()
	s.ClearWill()

	return nil
}

// called by the backend to queue an offline message
func (s *MemorySession) queue(msg *packet.Message) {
	s.offlineStore.Push(msg)
}

// called by the backend to retrieve all offline messages
func (s *MemorySession) nextMissed() *packet.Message {
	return s.offlineStore.Pop()
}

func (s *MemorySession) storeForDirection(dir Direction) *tools.Store {
	if dir == Incoming {
		return s.incStore
	} else if dir == Outgoing {
		return s.outStore
	}

	panic("unknown direction")
}
