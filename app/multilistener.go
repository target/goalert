package app

import (
	"net"
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

func (ml multiListener) Accept() (net.Conn, error) {
	select {
	case conn := <-ml.ch:
		return conn, nil
	case err := <-ml.errCh:
		return nil, err
	}
}

func (ml multiListener) Close() error {
	for _, l := range ml.listeners {
		err := l.Close()
		if err != nil {
			return err
		}

	}
	return nil
}

func (ml multiListener) Addr() net.Addr {
	//first addr in slice - might have to make better later
	return ml.listeners[0].Addr()
}
