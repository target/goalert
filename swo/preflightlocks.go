package swo

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/lock"
)

var ErrNoLock = errors.New("no lock")

// SwitchOverExecLock will attempt to grab the GlobalSwitchOverExec lock.
//
// After acquiring the lock, it will ensure the switchover has not yet been
// completed.
//
// This lock should be acquired by an engine instance that is going to perform
// the sync & switchover.
func SwitchOverExecLock(ctx context.Context, conn *pgx.Conn) error {
	var gotLock bool
	err := conn.QueryRow(ctx, `
		select pg_try_advisory_lock($1)
		from switchover_state
		where current_state != 'use_next_db'
	`, lock.GlobalSwitchOverExec).Scan(&gotLock)
	if err != nil {
		return err
	}

	if !gotLock {
		return ErrNoLock
	}

	return nil
}

// UnlockConn will release all session locks or close the connection.
func UnlockConn(ctx context.Context, conn *pgx.Conn) {
	_, err := conn.Exec(ctx, `select pg_advisory_unlock_all()`)
	if err != nil {
		conn.Close(ctx)
	}
}

var errDone = errors.New("done")

// sessionLock will get a shared advisory lock for the connection.
func sessionLock(ctx context.Context, conn driver.Conn) error {
	type execQuery interface {
		driver.ExecerContext
		driver.QueryerContext
	}

	c := conn.(execQuery)

	// Using literal here so we can avoid a prepared statement round trip.
	//
	// This will run for every new connection in SWO mode and for every
	// query while idle connections are disabled during critical phase.
	_, err := c.ExecContext(ctx, fmt.Sprintf("select pg_advisory_lock_shared(%d)", lock.GlobalSwitchOver), nil)
	if err != nil {
		return fmt.Errorf("get SWO shared session lock: %w", err)
	}

	rows, err := c.QueryContext(ctx, "select current_state from switchover_state", nil)
	if err != nil {
		return fmt.Errorf("get current SWO state: %w", err)
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
		return fmt.Errorf("get current SWO state: expected string for current_state value, got %t", t)
	}
	err = rows.Close()
	if err != nil {
		return err
	}

	if state == "use_next_db" {
		return errDone
	}

	return nil
}
