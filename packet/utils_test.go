package packet

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	t.Run("M4", func(t *testing.T) {
		fn(t, M4)
	})

	t.Run("M5", func(t *testing.T) {
		fn(t, M5)
	})
}

func assertDecodeError(t *testing.T, m Mode, typ Type, read int, packet []byte) {
	pkt, _ := typ.New()
	n, err := pkt.Decode(m, packet)
	assert.Error(t, err)
	assert.Equal(t, read, n)
}

func assertEncodeError(t *testing.T, m Mode, len, written int, pkt Generic) {
	if len == 0 {
		len = pkt.Len(m)
	}
	dst := make([]byte, len)
	n, err := pkt.Encode(m, dst)
	assert.Error(t, err)
	assert.Equal(t, written, n)
}

func benchTest(b *testing.B, fn func(b *testing.B, m Mode)) {
	b.ReportAllocs()

	b.Run("M4", func(b *testing.B) {
		fn(b, M4)
	})

	b.Run("M5", func(b *testing.B) {
		fn(b, M5)
	})
}

func benchPacket(b *testing.B, pkt Generic) {
	benchTest(b, func(b *testing.B, m Mode) {
		buf := make([]byte, pkt.Len(m))

		b.Run("Encode", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := pkt.Encode(m, buf)
				if err != nil {
					panic(err)
				}
			}
		})

		b.Run("Decode", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := pkt.Decode(m, buf)
				if err != nil {
					panic(err)
				}
			}
		})
	})
}
