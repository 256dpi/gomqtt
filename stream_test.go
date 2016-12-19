package packet

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)

	err := enc.Write(NewConnectPacket())
	assert.NoError(t, err)

	err = enc.Flush()
	assert.NoError(t, err)

	assert.Len(t, buf.Bytes(), 14)
}

func TestEncoderError1(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)

	pkt := NewConnackPacket()
	pkt.ReturnCode = 11 // <- invalid return code

	err := enc.Write(pkt)
	assert.Error(t, err)
}

func TestEncoderError2(t *testing.T) {
	enc := NewEncoder(&errorWriter{})

	pkt := NewPublishPacket()
	pkt.Message.Topic = "foo"
	pkt.Message.Payload = make([]byte, 4096)

	err := enc.Write(pkt)
	assert.Error(t, err)
}

func TestDecoder(t *testing.T) {
	buf := new(bytes.Buffer)
	dec := NewDecoder(buf)

	var pkt Packet = NewConnectPacket()
	b := make([]byte, pkt.Len())
	pkt.Encode(b)
	buf.Write(b)

	pkt, err := dec.Read()
	assert.NoError(t, err)
	assert.NotNil(t, pkt)
}

func TestStream(t *testing.T) {
	in := new(bytes.Buffer)
	out := new(bytes.Buffer)

	s := NewStream(in, out)

	err := s.Write(NewConnectPacket())
	assert.NoError(t, err)

	err = s.Flush()
	assert.NoError(t, err)

	_, err = io.Copy(in, out)
	assert.NoError(t, err)

	pkt, err := s.Read()
	assert.NotNil(t, pkt)
	assert.NoError(t, err)
}
