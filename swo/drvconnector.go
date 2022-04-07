package swo

import (
	"context"
	"database/sql/driver"
	"errors"
	"sync"

	"github.com/jackc/pgx/v4/stdlib"
	"github.com/target/goalert/swo/swogrp"
)

type Connector struct {
	dbcOld, dbcNew driver.Connector

	isDone bool
	id     int
	mx     sync.Mutex
}

var _ driver.Connector = (*Connector)(nil)

func NewConnector(dbcOld, dbcNew driver.Connector) *Connector {
	return &Connector{
		dbcOld: dbcOld,
		dbcNew: dbcNew,
	}
}

func (drv *Connector) Driver() driver.Driver { return nil }

func (drv *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	drv.mx.Lock()
	isDone := drv.isDone
	drv.mx.Unlock()

	if isDone {
		return drv.dbcNew.Connect(ctx)
	}

	c, err := drv.dbcOld.Connect(ctx)
	if err != nil {
		return nil, err
	}

	drv.id++
	conn := c.(*stdlib.Conn)

	err = sessionLock(ctx, conn)
	if errors.Is(err, swogrp.ErrDone) {
		drv.mx.Lock()
		drv.isDone = true
		drv.mx.Unlock()
		return drv.dbcNew.Connect(ctx)
	}
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
