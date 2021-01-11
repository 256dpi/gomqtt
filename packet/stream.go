package packet

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/256dpi/mercury"
)

// ErrDetectionOverflow is returned by the Decoder if the next packet couldn't
// be detect from the initial header bytes.
var ErrDetectionOverflow = errors.New("detection overflow")

// ErrReadLimitExceeded can be returned during a Receive if the connection
// exceeded its read limit.
var ErrReadLimitExceeded = errors.New("read limit exceeded")

// An Encoder wraps a writer and continuously encodes packets.
type Encoder struct {
	mode   Mode
	writer *mercury.Writer
	buffer bytes.Buffer
	mutex  sync.Mutex
}

// NewEncoder creates a new Encoder.
func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{
		writer: mercury.NewWriter(writer, 0),
	}
}

// SetMode will set the encoding mode.
func (e *Encoder) SetMode(m Mode) {
	// acquire mutex
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// set mode
	e.mode = m
}

// Write encodes and writes the passed packet to the write buffer.
func (e *Encoder) Write(pkt Generic, async bool) error {
	// acquire mutex
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// reset and potentially grow buffer
	packetLength := pkt.Len(e.mode)
	e.buffer.Reset()
	e.buffer.Grow(packetLength)
	buf := e.buffer.Bytes()[0:packetLength]

	// encode packet
	_, err := pkt.Encode(e.mode, buf)
	if err != nil {
		return err
	}

	// write buffer
	if async {
		_, err = e.writer.Write(buf)
	} else {
		_, err = e.writer.WriteAndFlush(buf)
	}
	if err != nil {
		return err
	}

	return nil
}

// Flush flushes the writer buffer.
func (e *Encoder) Flush() error {
	return e.writer.Flush()
}

// SetMaxWriteDelay will set the maximum amount of time allowed to pass until
// an asynchronous write is flushed.
func (e *Encoder) SetMaxWriteDelay(delay time.Duration) {
	e.writer.SetMaxDelay(delay)
}

// A Decoder wraps a Reader and continuously decodes packets.
type Decoder struct {
	limit  int64
	mode   Mode
	reader *bufio.Reader
	buffer bytes.Buffer
	mutex  sync.Mutex
}

// NewDecoder returns a new Decoder.
func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(reader),
	}
}

// SetMode will set the decoding mode.
func (d *Decoder) SetMode(m Mode) {
	// acquire mutex
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// set mode
	d.mode = m
}

// Read reads the next packet from the buffered reader.
func (d *Decoder) Read() (Generic, error) {
	// acquire mutex
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// initial detection length
	detectionLength := 2

	for {
		// check length
		if detectionLength > 5 {
			return nil, ErrDetectionOverflow
		}

		// try read detection bytes
		header, err := d.reader.Peek(detectionLength)
		if err == io.EOF && len(header) != 0 {
			// an EOF with some data is unexpected
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}

		// detect packet
		packetLength, packetType := DetectPacket(header)

		// on zero packet length:
		// increment detection length and try again
		if packetLength <= 0 {
			detectionLength++
			continue
		}

		// check read limit
		limit := atomic.LoadInt64(&d.limit)
		if limit > 0 && int64(packetLength) > limit {
			return nil, ErrReadLimitExceeded
		}

		// create packet
		pkt, err := packetType.New()
		if err != nil {
			return nil, err
		}

		// reset and eventually grow buffer
		d.buffer.Reset()
		d.buffer.Grow(packetLength)
		buf := d.buffer.Bytes()[0:packetLength]

		// read whole packet (will not return EOF)
		_, err = io.ReadFull(d.reader, buf)
		if err != nil {
			return nil, err
		}

		// decode buffer
		_, err = pkt.Decode(d.mode, buf)
		if err != nil {
			return nil, err
		}

		return pkt, nil
	}
}

// SetReadLimit will set the read limit. Packets with a length above that limit
// will cause the ErrReadLimitExceeded error.
func (d *Decoder) SetReadLimit(limit int64) {
	atomic.StoreInt64(&d.limit, limit)
}

// A Stream combines an Encoder and Decoder
type Stream struct {
	*Decoder
	*Encoder
}

// NewStream creates a new Stream.
func NewStream(reader io.Reader, writer io.Writer) *Stream {
	return &Stream{
		Decoder: NewDecoder(reader),
		Encoder: NewEncoder(writer),
	}
}

// SetMode will set the coding mode.
func (s *Stream) SetMode(m Mode) {
	s.Decoder.SetMode(m)
	s.Encoder.SetMode(m)
}
