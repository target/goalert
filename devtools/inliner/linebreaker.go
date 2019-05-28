package main

import (
	"io"
)

const lineLength = 80

type lineBreaker struct {
	pos int

	out io.Writer
}

var nl = []byte{'\n'}

func (l *lineBreaker) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if l.pos+len(b) < lineLength {
		l.pos += len(b)
		return l.out.Write(b)
	}

	diff := lineLength - l.pos

	n, err = l.out.Write(b[:diff])
	if err != nil {
		return n, err
	}
	l.pos = 0

	n, err = l.out.Write(nl)
	if err != nil {
		return diff + n, err
	}

	n, err = l.Write(b[diff:])
	return lineLength + 1 + n, err
}

func (l *lineBreaker) Close() (err error) {
	if l.pos > 0 {
		_, err = l.out.Write(nl)
	}
	return err
}
