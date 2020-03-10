package sendit

import (
	"io"
	"net/http"
)

type flushWriter struct {
	io.Writer
	flush func()
}

// FlushWriter will wrap an `io.Writer`, calling `.Flush` after each `.Write` call
// if it implements the `http.Flusher` interface.
func FlushWriter(w io.Writer) io.Writer {
	if f, ok := w.(http.Flusher); ok {
		return &flushWriter{Writer: w, flush: f.Flush}
	}
	return w
}
func (w *flushWriter) Write(p []byte) (int, error) {
	defer w.flush()
	return w.Writer.Write(p)
}
