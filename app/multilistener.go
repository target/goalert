package app

import (
	"errors"
	"net"
	"sync"
)

type multiListener struct {
	listeners []net.Listener

	ch      chan net.Conn
	errCh   chan error
	closeCh chan struct{}
	closed  bool
	wg      sync.WaitGroup
}

func newMultiListener(ln ...net.Listener) *multiListener {
	nonEmpty := make([]net.Listener, 0, len(ln))
	for _, l := range ln {
		if l != nil {
			nonEmpty = append(nonEmpty, l)
		}
	}
	ln = nonEmpty

	ml := multiListener{
		listeners: ln,
		ch:        make(chan net.Conn),
		errCh:     make(chan error),
		closeCh:   make(chan struct{}),
	}
	for _, l := range ln {
		ml.wg.Add(1)
		go ml.listen(l)
	}
	return &ml
}

// listen waits for and returns the next connection for the listener.
func (ml *multiListener) listen(l net.Listener) {
	defer ml.wg.Done()
	for {
		c, err := l.Accept()
		if err != nil {
			select {
			case ml.errCh <- err:
			case <-ml.closeCh:
				return
			}
			return
		}
		select {
		case ml.ch <- c:
		case <-ml.closeCh:
			c.Close()
			return
		}
	}
}

// Accept retrieves the contents from the connection and error channels of the multilistener.
// Based on the results, either the next connection is returned or the error.
func (ml *multiListener) Accept() (net.Conn, error) {
	select {
	case conn := <-ml.ch:
		return conn, nil
	case err := <-ml.errCh:
		return nil, err
	case <-ml.closeCh:
		return nil, errors.New("listener is closed")
	}
}

// Close ranges through listeners closing all of them and and returns an error if any listener encountered an error while closing.
func (ml *multiListener) Close() error {
	defer ml.wg.Wait()
	if !ml.closed {
		close(ml.closeCh)
		ml.closed = true
	}

	var errs []error
	for _, l := range ml.listeners {
		err := l.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Addr returns the address of the first listener in the multilistener.
// This implementation of Addr might change in the future.
func (ml *multiListener) Addr() net.Addr {
	return ml.listeners[0].Addr()
}
