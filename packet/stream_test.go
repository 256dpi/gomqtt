package packet

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	enc.SetMaxWriteDelay(time.Millisecond)

	err := enc.Write(NewConnect(), false)
	assert.NoError(t, err)

	err = enc.Flush()
	assert.NoError(t, err)

	assert.Len(t, buf.Bytes(), 14)
}

func TestEncoderEncodeError(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	enc.SetMaxWriteDelay(time.Millisecond)

	pkt := NewConnack()
	pkt.ReturnCode = 11 // < invalid return code

	err := enc.Write(pkt, false)
	assert.Error(t, err)
}

func TestEncoderWriterError(t *testing.T) {
	enc := NewEncoder(&errorWriter{
		err: errors.New("foo"),
	})
	enc.SetMaxWriteDelay(time.Millisecond)

	pkt := NewPublish()
	pkt.Message.Topic = "foo"
	pkt.Message.Payload = make([]byte, 4096)

	err := enc.Write(pkt, false)
	assert.Error(t, err)
}

func TestDecoder(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)

	var pkt Generic = NewConnect()
	b := make([]byte, pkt.Len(M4))
	_, err := pkt.Encode(M4, b)
	assert.NoError(t, err)
	buf.Write(b)

	pkt, err = dec.Read()
	assert.NoError(t, err)
	assert.NotNil(t, pkt)
}

func TestDecoderDetectionOverflowError(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)

	buf.Write([]byte{0x10, 0xff, 0xff, 0xff, 0x80})

	pkt, err := dec.Read()
	assert.Equal(t, ErrDetectionOverflow, err)
	assert.Nil(t, pkt)
}

func TestDecoderPeekUnexpectedEOFError(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)

	buf.Write([]byte{0x10, 0xff, 0xff})

	pkt, err := dec.Read()
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.Nil(t, pkt)
}

func TestDecoderPeekError(t *testing.T) {
	dec := NewDecoder(&errorReader{
		err: errors.New("foo"),
	})

	pkt, err := dec.Read()
	assert.Error(t, err)
	assert.Nil(t, pkt)
}

func TestDecoderInvalidTypeError(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)

	buf.Write([]byte{0x00, 0x00})

	pkt, err := dec.Read()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid packet type")
	assert.Nil(t, pkt)
}

func TestDecoderReadLimitError(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)
	dec.limit = 1

	buf.Write([]byte{0x00, 0x00})

	pkt, err := dec.Read()
	assert.Equal(t, ErrReadLimitExceeded, err)
	assert.Nil(t, pkt)
}

func TestDecoderReadError(t *testing.T) {
	dec := NewDecoder(&errorReader{
		reader: bytes.NewBuffer([]byte{0x10, 0xc, 0x0, 0x4}),
		after:  1,
		err:    errors.New("foo"),
	})

	pkt, err := dec.Read()
	assert.Error(t, err)
	assert.Nil(t, pkt)
}

func TestDecoderDecodeError(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)

	buf.Write([]byte{0x20, 0x02, 0x40, 0x00})

	pkt, err := dec.Read()
	assert.Error(t, err)
	assert.Nil(t, pkt)
}

func TestStream(t *testing.T) {
	in := new(bytes.Buffer)
	out := new(bytes.Buffer)

	s := NewStream(in, out)
	s.SetMaxWriteDelay(time.Millisecond)

	err := s.Write(NewConnect(), false)
	assert.NoError(t, err)

	err = s.Flush()
	assert.NoError(t, err)

	_, err = io.Copy(in, out)
	assert.NoError(t, err)

	pkt, err := s.Read()
	assert.NotNil(t, pkt)
	assert.NoError(t, err)
}
