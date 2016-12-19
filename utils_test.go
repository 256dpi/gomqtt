package packet

import "errors"

type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("foo")
}
