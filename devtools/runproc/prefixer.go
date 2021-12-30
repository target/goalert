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

		err := w.writePrefix()
		if err != nil {
			return n, err
		}

		if len(w.buf) > 0 {
			_, err = w.out.Write(w.buf)
			if err != nil {
				return n, err
			}
			w.buf = w.buf[:0]
		}

		_n, err := w.out.Write(p[:l+1])
		n += _n
		if err != nil {
			return n, err
		}

		p = p[l+1:]
	}
}
