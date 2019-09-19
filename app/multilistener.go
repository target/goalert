package app

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
)

type multiListener struct {
	listeners []net.Listener

	ch    chan net.Conn
	errCh chan error
}

func newMultiListener(ln ...net.Listener) *multiListener {
	ml := multiListener{listeners: ln}
	ml.ch = make(chan net.Conn)
	ml.errCh = make(chan error)

	for _, l := range ln {
		go ml.listen(l)
	}
	return &ml
}

// listen waits for and returns the next connection for the listener.
func (ml multiListener) listen(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			ml.errCh <- err
			return
		}
		ml.ch <- c
	}
}

// Accept retrieves the contents from the connection and error channels of the multilistener.
// Based on the results, either the next connection is returned or the error.
func (ml multiListener) Accept() (net.Conn, error) {
	select {
	case conn := <-ml.ch:
		return conn, nil
	case err := <-ml.errCh:
		return nil, err
	}
}

// Close ranges through listeners closing all of them and and returns an error if any listener encountered an error while closing.
// It will log all individual listener errors and return an error in the end in the case of error(s).
func (ml multiListener) Close() error {
	hasErr := false
	for _, l := range ml.listeners {
		err := l.Close()
		if err != nil {
			hasErr = true
			log.Log(context.Background(), errors.Wrap(err, "Listener error "))
		}
	}

	if hasErr {
		return errors.New("Multiple listeners failed.")
	}
	return nil
}

// Addr returns the address of the first listener in the multilistener.
// This implementation of Addr might change in the future.
func (ml multiListener) Addr() net.Addr {
	return ml.listeners[0].Addr()
}
