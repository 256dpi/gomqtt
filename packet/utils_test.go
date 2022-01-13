package packet

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var longBytes = bytes.Repeat([]byte{'X'}, 65536)
var longString = string(longBytes)

type errorWriter struct {
	writer io.Writer
	after  int
	err    error
}

func (w *errorWriter) Write(p []byte) (int, error) {
	if w.after == 0 {
		return 0, w.err
	}

	return w.writer.Write(p)
}

type errorReader struct {
	reader io.Reader
	after  int
	err    error
}

func (r *errorReader) Read(p []byte) (int, error) {
	if r.after == 0 {
		return 0, r.err
	}

	r.after--
	return r.reader.Read(p)
}

func multiTest(t *testing.T, fn func(t *testing.T, m Mode)) {
	t.Run("M3", func(t *testing.T) {
		fn(t, M3)
	})

	t.Run("M3L", func(t *testing.T) {
		fn(t, M3L)
	})

	t.Run("M4", func(t *testing.T) {
		fn(t, M4)
	})

	t.Run("M4L", func(t *testing.T) {
		fn(t, M4L)
	})

	t.Run("M5", func(t *testing.T) {
		fn(t, M5)
	})

	t.Run("M5L", func(t *testing.T) {
		fn(t, M5L)
	})
}

func assertEncodeDecode(t *testing.T, m Mode, packet []byte, pkt Generic, str string) {
	out, _ := pkt.Type().New()
	n, err := out.Decode(m, packet)
	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, pkt, out)

	buf := make([]byte, pkt.Len(m))
	n, err = pkt.Encode(m, buf)
	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
	assert.Equal(t, packet, buf)

	if str != "" {
		assert.Equal(t, str, pkt.String())
	}
}

func assertDecodeError(t *testing.T, m Mode, typ Type, read int, packet []byte, r interface{}) {
	var e Error
	e.Type = typ
	e.Method = DECODE
	e.Mode = m
	e.Position = read
	switch r := r.(type) {
	case string:
		e.Message = r
	case error:
		e.Err = r
	}

	pkt, _ := typ.New()
	n, err := pkt.Decode(m, packet)
	assert.Error(t, err)
	assert.Equal(t, read, n)
	assert.Equal(t, &e, err)
}

func assertEncodeError(t *testing.T, m Mode, len, written int, pkt Generic, r interface{}) {
	var e Error
	e.Type = pkt.Type()
	e.Method = ENCODE
	e.Mode = m
	e.Position = written
	switch r := r.(type) {
	case string:
		e.Message = r
	case error:
		e.Err = r
	}

	if len == 0 {
		len = pkt.Len(m)
	}

	dst := make([]byte, len)
	n, err := pkt.Encode(m, dst)
	assert.Error(t, err)
	assert.Equal(t, written, n)
	assert.Equal(t, &e, err)
}

func multiBench(b *testing.B, fn func(b *testing.B, m Mode)) {
	b.Run("M4", func(b *testing.B) {
		fn(b, M4)
	})

	b.Run("M5", func(b *testing.B) {
		fn(b, M5)
	})
}

func benchPacket(b *testing.B, pkt Generic) {
	multiBench(b, func(b *testing.B, m Mode) {
		buf := make([]byte, pkt.Len(m))

		b.Run("ENCODE", func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := pkt.Encode(m, buf)
				if err != nil {
					panic(err)
				}
			}
		})

		b.Run("Decode", func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := pkt.Decode(m, buf)
				if err != nil {
					panic(err)
				}
			}
		})
	})
}
