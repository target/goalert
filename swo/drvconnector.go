package swo

import (
	"context"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/target/goalert/version"
)

type Connector struct {
	dbcOld, dbcNew driver.Connector

	isDone bool
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

	conn := c.(*stdlib.Conn)

	var b pgx.Batch
	b.Queue("select pg_advisory_lock_shared(4369)")
	b.Queue("select current_state = 'use_next_db' FROM switchover_state")

	res := conn.Conn().SendBatch(ctx, &b)
	if _, err := res.Exec(); err != nil {
		conn.Close()
		return nil, err
	}
	defer res.Close()

	var useNext bool
	if err := res.QueryRow().Scan(&useNext); err != nil {
		conn.Close()
		return nil, err
	}

	if useNext {
		conn.Close()
		drv.mx.Lock()
		drv.isDone = true
		drv.mx.Unlock()
		return drv.dbcNew.Connect(ctx)
	}

	return conn, nil
}
