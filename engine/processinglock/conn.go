package processinglock

import (
	"context"
	"database/sql"
	"sync"

	"github.com/target/goalert/util/sqlutil"
)

// Conn allows using locked transactions over a single connection.
type Conn struct {
	l    *Lock
	conn *sql.Conn
	mx   sync.Mutex
}

// Conn returns a new connection from the DB pool.
//
// Note: No version checking/locking is done until a transaction is started.
func (l *Lock) Conn(ctx context.Context) (*Conn, error) {
	c, err := l.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	_, err = c.ExecContext(ctx, `SET idle_in_transaction_session_timeout = 3000`)
	if err != nil {
		c.Close()
		return nil, err
	}

	_, err = c.ExecContext(ctx, `SET lock_timeout = 8000`)
	if err != nil {
		c.Close()
		return nil, err
	}

	return &Conn{l: l, conn: c}, nil
}

// BeginTx will start a new transaction.
func (c *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.l._BeginTx(ctx, c.conn, opts, false)
}

// WithTx will run the given function in a locked transaction.
func (c *Conn) WithTx(ctx context.Context, txFn func(tx *sql.Tx) error) error {
	c.mx.Lock()
	defer c.mx.Unlock()
	tx, err := c.l._BeginTx(ctx, c.conn, nil, false)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "rollback tx", tx)

	err = txFn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Exec will call ExecContext on the statement wrapped in a locked transaction.
func (c *Conn) Exec(ctx context.Context, stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.l._Exec(ctx, c.conn, stmt, args...)
}

// ExecWithoutLock will run a query directly on the connection (no Tx or locking).
func (c *Conn) ExecWithoutLock(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.conn.ExecContext(ctx, query, args...)
}

// Close returns the connection to the pool.
func (c *Conn) Close() error { return c.conn.Close() }
