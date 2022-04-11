package main

import (
	"io"
	"math/rand"
	"time"
)

type DelayWriter struct {
	io.Writer
	latency time.Duration
	jitter  time.Duration
}

func (w *DelayWriter) Write(p []byte) (int, error) {
	delay := w.latency - (w.jitter / 2) + time.Duration(rand.Float64()*float64(w.jitter))
	time.Sleep(delay)

	return w.Writer.Write(p)
}
