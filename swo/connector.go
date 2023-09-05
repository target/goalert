package swo

import (
	"context"
	"database/sql/driver"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Connector is a driver.Connector that will use the old database until the
// switchover_state table indicates that the new database should be used.
//
// Until the switchover is complete, the old database will be protected with a
// shared advisory lock (4369).
type Connector struct {
	dbcOld, dbcNew driver.Connector

	isDone bool
	mx     sync.Mutex
}

var _ driver.Connector = (*Connector)(nil)

// NewConnector creates a new Connector pointing to the old and new DBs, respectively.
func NewConnector(dbcOld, dbcNew driver.Connector) *Connector {
	return &Connector{
		dbcOld: dbcOld,
		dbcNew: dbcNew,
	}
}

// Driver is a stub method for driver.Connector that returns nil.
func (drv *Connector) Driver() driver.Driver { return nil }

// Connect returns a new connection to the database.
//
// A shared advisory lock is acquired on the connection to the old DB, and then switchover_state is checked.
//
// When `current_state` is `use_next_db`, the connection is closed and a connection to the new DB is returned
// instead. Future connections then skip the check and are returned directly from the new DB.
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
