package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var mx sync.Mutex

type prefixWriter struct {
	prefix string
	io.Writer
	stop bool
}

var blockOnce sync.Once

func (w *prefixWriter) Write(p []byte) (int, error) {
	if w.stop {
		return len(p), nil
	}
	mx.Lock()
	fmt.Println(w.prefix + "\n\t" + strings.ReplaceAll(hex.Dump(p), "\n", "\n\t"))
	mx.Unlock()

	if drop != "" && bytes.Contains(p, []byte(drop)) {
		if dropFirst {
			blockOnce.Do(func() {
				w.stop = true
				time.Sleep(time.Minute)
			})
		} else {
			w.stop = true
			time.Sleep(time.Minute)
		}

	}
	return w.Writer.Write(p)
}
