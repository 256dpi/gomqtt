package packet

import (
	"errors"
	"io"
)

type errorWriter struct {
	writer io.Writer
	after  int
}

func (w *errorWriter) Write(p []byte) (int, error) {
	if w.after == 0 {
		return 0, errors.New("foo")
	}

	return w.writer.Write(p)
}

type errorReader struct {
	reader io.Reader
	after  int
}

func (r *errorReader) Read(p []byte) (int, error) {
	if r.after == 0 {
		return 0, errors.New("foo")
	}

	r.after--
	return r.reader.Read(p)
}
