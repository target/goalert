package app

import (
	"net"

	"github.com/pkg/errors"
)

func listenStatus(addr string, done <-chan struct{}) error {
	if addr == "" {
		return nil
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "start status listener")
	}
	ch := make(chan net.Conn)

	go func() {
		defer close(ch)
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			ch <- c
		}
	}()
	go func() {
		var conn []net.Conn
	loop:
		for {
			select {
			case <-done:
				l.Close()
				break loop
			case c := <-ch:
				conn = append(conn, c)
			}
		}
		for c := range ch {
			c.Close()
		}
		for _, c := range conn {
			c.Close()
		}
	}()

	return nil
}
