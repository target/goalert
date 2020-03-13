package sendit

import (
	"io"
	"net/http"
)

type flushWriter struct {
	write io.Writer
	flush func()
}

func (f *flushWriter) Write(p []byte) (n int, err error) {
	n, err = f.write.Write(p)
	if err == nil {
		f.flush()
	}
	return n, err
}

// FlushWriter will call .Flush() immediately after every call to .Write() on the returned io.Writer.
//
// If w does not implement http.Flusher, it panics.
func FlushWriter(w io.Writer) io.Writer {
	flush := w.(http.Flusher).Flush
	flush()
	return &flushWriter{
		write: w,
		flush: flush,
	}
}
