package main

import (
	"io"
	"math/rand"
	"time"
)

type rateLimiter struct {
	bucket   chan int
	overflow chan int
	rate     bool
	latency  time.Duration
	jitter   time.Duration
}

func newRateLimiter(bps int, latency, jitter time.Duration) *rateLimiter {
	ch := make(chan int)
	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			ch <- bps
		}
	}()
	return &rateLimiter{
		rate:     bps > 0,
		bucket:   ch,
		overflow: make(chan int, 1000),
		latency:  latency,
		jitter:   jitter,
	}
}

func (r *rateLimiter) WaitFor(count int) time.Duration {
	var n int
	for r.rate && n < count {
		select {
		case val := <-r.bucket:
			n += val
		case val := <-r.overflow:
			n += val
		}
	}
	if n > count {
		r.overflow <- n - count
	}
	return (r.latency - (r.jitter / 2) + time.Duration(rand.Float64()*float64(r.jitter))) / 2
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
