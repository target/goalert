package main

import (
	"io"
	"math/rand"
	"time"
)

type rateLimiter struct {
	bucket   chan int
	overflow chan int
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
		bucket:   ch,
		overflow: make(chan int, 1000),
		latency:  latency,
		jitter:   jitter,
	}
}

func (r *rateLimiter) WaitFor(count int) {
	delay := r.latency - (r.jitter / 2) + time.Duration(rand.Float64()*float64(r.jitter))
	s := time.Now()
	var n int
	for n < count {
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
	time.Sleep(delay - time.Since(s))
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
	w.l.WaitFor(len(p))
	return w.w.Write(p)
}
