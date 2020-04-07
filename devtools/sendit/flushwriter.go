package sendit

import (
	"io"
	"log"
	"net/http"
)

type flushWriter struct {
	write io.Writer
	flush func()
}

func (f *flushWriter) Write(p []byte) (n int, err error) {
	n, err = f.write.Write(p)
	if err == nil {
		defer func() {
			err := recover()
			if err != nil {
				// .flush() calls Flush() on the http.ResponseWriter
				// However, if the connection errs, the underlying connection
				// can be closed before we return the handlers goroutine.
				//
				// This means it's possible for the Flush() call to panic as
				// the `finalFlush` sets bufw to nil.
				log.Println("ERROR:", err)
			}
		}()
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
