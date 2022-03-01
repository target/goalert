package swo

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"
)

type Conn struct {
	DBConn

	n      *Notifier
	locked bool
}

var ErrDone = fmt.Errorf("switchover is already done")

type DBConn interface {
	driver.Conn
	driver.Pinger
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ConnBeginTx
	driver.NamedValueChecker
}

var (
	_ driver.SessionResetter = (*Conn)(nil)
	_ driver.Validator       = (*Conn)(nil)
)

func (c *Conn) lock(ctx context.Context) error {
	if c.n.IsDone() {
		return driver.ErrBadConn
	}
	if c.locked {
		return nil
	}

	_, err := c.ExecContext(ctx, "select pg_advisory_lock_shared(4369)", nil)
	if err != nil {
		return err
	}
	c.locked = true

	rows, err := c.QueryContext(ctx, "select current_state from switchover_state", nil)
	if err != nil {
		return err
	}

	scan := make([]driver.Value, 1)
	err = rows.Next(scan)
	if err != nil {
		return err
	}

	var state string
	switch t := scan[0].(type) {
	case string:
		state = t
	case []byte:
		state = string(t)
	default:
		return fmt.Errorf("expected string for current_state value, got %t", t)
	}
	err = rows.Close()
	if err != nil {
		return err
	}

	if state == "use_next_db" {
		c.n.Done()
		return driver.ErrBadConn
	}

	return nil
}

func (c *Conn) unlock(ctx context.Context) error {
	if !c.locked {
		return nil
	}

	_, err := c.ExecContext(ctx, "select pg_advisory_unlock_shared(4369)", nil)
	if err != nil {
		return err
	}

	c.locked = false

	return nil
}

func (c *Conn) ResetSession(ctx context.Context) error {
	err := c.lock(ctx)
	if err != nil {
		return err
	}

	if s, ok := c.DBConn.(driver.SessionResetter); ok {
		return s.ResetSession(ctx)
	}

	return nil
}

func (c *Conn) IsValid() bool {
	if c.n.IsDone() {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := c.unlock(ctx); err != nil {
		return false
	}

	return true
}
