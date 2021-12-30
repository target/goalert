package main

import (
	"bytes"
	"io"
)

type prefixer struct {
	prefix string
	out    io.Writer
	buf    []byte
}

func NewPrefixer(out io.Writer, prefix string) io.Writer {
	return &prefixer{
		out:    out,
		prefix: prefix,
	}
}

func (w *prefixer) writePrefix() error {
	_, err := io.WriteString(w.out, w.prefix+" | ")
	return err
}

func (w *prefixer) Write(p []byte) (int, error) {
	var n int
	for {
		l := bytes.IndexByte(p, '\n')
		if l == -1 {
			w.buf = append(w.buf, p...)
			return n + len(p), nil
		}
		w.buf = append(w.buf, p[:l+1]...)
		n += l + 1

		err := w.writePrefix()
		if err != nil {
			return n, err
		}

		// replace yarn escape sequences
		w.buf = bytes.ReplaceAll(w.buf, []byte("\x1b[2K"), nil)
		w.buf = bytes.ReplaceAll(w.buf, []byte("\x1b[1G"), nil)

		_, err = w.out.Write(w.buf)
		if err != nil {
			return n, err
		}
		w.buf = w.buf[:0]

		p = p[l+1:]
	}
}
