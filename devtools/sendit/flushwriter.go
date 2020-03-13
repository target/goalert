package sendit

import (
	"context"
	"io"
	"net/http"
	"time"
)

// FlushWriter will spawn a goroutine that will constantly flush the writer every
// delay interval. It exits when the context expires.
//
// If w does not implement http.Flusher, it panics.
func FlushWriter(ctx context.Context, w io.Writer, delay time.Duration) {
	f := w.(http.Flusher)
	go func() {
		t := time.NewTicker(delay)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				f.Flush()
			}
		}
	}()
}
