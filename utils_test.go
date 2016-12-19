package packet

import "io"

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
