package main

import (
	"io"
	"math/rand"
	"time"
)

type rateLimiter struct {
	bucket  chan int
	rate    bool
	latency time.Duration
	jitter  time.Duration
}

func newRateLimiter(bps int, latency, jitter time.Duration) *rateLimiter {
	ch := make(chan int)

	bpMs := float64(bps) / 1000
	go func() {
		t := time.NewTicker(time.Millisecond)
		var count float64
		for {
			if count >= bpMs {
				<-t.C
				count -= bpMs
				if count < 0 {
					count = 0
				}
				continue
			}

			select {
			case <-t.C:
				count -= bpMs
				if count < 0 {
					count = 0
				}
			case val := <-ch:
				count += float64(val)
			}
		}
	}()

	return &rateLimiter{
		rate:    bps > 0,
		bucket:  ch,
		latency: latency,
		jitter:  jitter,
	}
}

func (r *rateLimiter) WaitFor(count int) time.Duration {
	waitUntil := time.Now().Add((r.latency - (r.jitter / 2) + time.Duration(rand.Float64()*float64(r.jitter))))
	r.bucket <- count
	return time.Until(waitUntil) / 2
}

func (r *rateLimiter) NewWriter(w io.Writer) io.Writer {
	return &rateLimitWriter{
		w: w,
		l: r,
	}
}

type rateLimitWriter struct {
	l *rateLimiter
	w io.Writer
}

func (w *rateLimitWriter) Write(p []byte) (int, error) {
	dur := w.l.WaitFor(len(p))
	time.Sleep(dur)
	defer time.Sleep(dur)
	return w.w.Write(p)
}
