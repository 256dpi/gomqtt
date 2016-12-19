package transport

import (
	"io"
	"sync"
	"time"

	"github.com/gomqtt/packet"
)

// TODO: Expose conn and stream?

type carrier interface {
	io.ReadWriteCloser

	SetReadDeadline(time.Time) error
}

type Stream struct {
	carrier carrier

	stream *packet.Stream

	flushTimer *time.Timer
	flushError error

	sMutex sync.Mutex
	rMutex sync.Mutex

	readTimeout time.Duration
}

// Send will write the packet to the underlying connection. It will return
// an Error if there was an error while encoding or writing to the
// underlying connection.
//
// Note: Only one goroutine can Send at the same time.
func (s *Stream) Send(pkt packet.Packet) error {
	s.sMutex.Lock()
	defer s.sMutex.Unlock()

	// write packet
	err := s.write(pkt)
	if err != nil {
		return err
	}

	// stop the timer if existing
	if s.flushTimer != nil {
		s.flushTimer.Stop()
	}

	// flush buffer
	return s.flush()
}

// BufferedSend will write the packet to an internal buffer. It will flush
// the internal buffer automatically when it gets stale. Encoding errors are
// directly returned as in Send, but any network errors caught while flushing
// the buffer at a later time will be returned on the next call.
//
// Note: Only one goroutine can call BufferedSend at the same time.
func (s *Stream) BufferedSend(pkt packet.Packet) error {
	s.sMutex.Lock()
	defer s.sMutex.Unlock()

	// create the timer if missing
	if s.flushTimer == nil {
		s.flushTimer = time.AfterFunc(flushTimeout, s.asyncFlush)
		s.flushTimer.Stop()
	}

	// return any error from asyncFlush
	if s.flushError != nil {
		return s.flushError
	}

	// write packet
	err := s.write(pkt)
	if err != nil {
		return err
	}

	// queue asyncFlush
	s.flushTimer.Reset(flushTimeout)

	return nil
}

func (s *Stream) write(pkt packet.Packet) error {
	err := s.stream.Write(pkt)
	if err != nil {
		// ensure connection gets closed
		s.carrier.Close()

		return err
	}

	return nil
}

func (s *Stream) flush() error {
	err := s.stream.Flush()
	if err != nil {
		// ensure connection gets closed
		s.carrier.Close()

		return err
	}

	return nil
}

func (s *Stream) asyncFlush() {
	s.sMutex.Lock()
	defer s.sMutex.Unlock()

	// flush buffer and save an eventual error
	err := s.flush()
	if err != nil {
		s.flushError = err
	}
}

// Receive will read from the underlying connection and return a fully read
// packet. It will return an Error if there was an error while decoding or
// reading from the underlying connection.
//
// Note: Only one goroutine can Receive at the same time.
func (s *Stream) Receive() (packet.Packet, error) {
	s.rMutex.Lock()
	defer s.rMutex.Unlock()

	// read next packet
	pkt, err := s.stream.Read()
	if err != nil {
		s.carrier.Close()
		return nil, err
	}

	// reset timeout
	s.resetTimeout()

	return pkt, nil
}

// Close will close the underlying connection and cleanup resources. It will
// return an Error if there was an error while closing the underlying
// connection.
func (s *Stream) Close() error {
	s.sMutex.Lock()
	defer s.sMutex.Unlock()

	err := s.carrier.Close()
	if err != nil {
		return err
	}

	return nil
}

// SetReadLimit sets the maximum size of a packet that can be received.
// If the limit is greater than zero, Receive will close the connection and
// return an Error if receiving the next packet will exceed the limit.
func (s *Stream) SetReadLimit(limit int64) {
	s.stream.Decoder.Limit = limit
}

// SetReadTimeout sets the maximum time that can pass between reads.
// If no data is received in the set duration the connection will be closed
// and Read returns an error.
func (s *Stream) SetReadTimeout(timeout time.Duration) {
	s.readTimeout = timeout
	s.resetTimeout()
}

func (s *Stream) resetTimeout() {
	if s.readTimeout > 0 {
		s.carrier.SetReadDeadline(time.Now().Add(s.readTimeout))
	} else {
		s.carrier.SetReadDeadline(time.Time{})
	}
}
