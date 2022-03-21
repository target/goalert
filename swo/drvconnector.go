package swo

import (
	"context"
	"database/sql/driver"
	"errors"
	"sync"
)

type Connector struct {
	dbcOld, dbcNew driver.Connector

	n  *Notifier
	sm *StatsManager
}

type Notifier struct {
	doneCh chan struct{}
	done   sync.Once
}

func NewNotifier() *Notifier {
	return &Notifier{
		doneCh: make(chan struct{}),
	}
}
func (n *Notifier) Done() { n.done.Do(func() { close(n.doneCh) }) }
func (n *Notifier) IsDone() bool {
	select {
	case <-n.doneCh:
		return true
	default:
		return false
	}
}

var _ driver.Connector = (*Connector)(nil)

func NewConnector(dbcOld, dbcNew driver.Connector, sm *StatsManager) *Connector {
	return &Connector{
		dbcOld: dbcOld,
		dbcNew: dbcNew,
		n:      NewNotifier(),
		sm:     sm,
	}
}

func (drv *Connector) Driver() driver.Driver { return nil }

func (drv *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	if drv.n.IsDone() {
		return drv.dbcNew.Connect(ctx)
	}

	conn, err := drv.dbcOld.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = sessionLock(ctx, conn)
	if errors.Is(err, errDone) {
		drv.n.Done()
		return drv.dbcNew.Connect(ctx)
	}
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
