package swo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swogrp"
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
	gotLock, err := swodb.New(conn).GlobalSwitchoverExecLock(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return swogrp.ErrDone
	}
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
	err := swodb.New(conn).UnlockAll(ctx)
	if err != nil {
		conn.Close(ctx)
	}
}

// SessionLock will get a shared advisory lock for the connection.
func SessionLock(ctx context.Context, c *stdlib.Conn) error {
	// Using literal here so we can avoid a prepared statement round trip.
	//
	// This will run for every new connection in SWO mode and for every
	// query while idle connections are disabled during critical phase.
	err := swodb.New(c.Conn()).GlobalSwitchoverSharedConnLock(ctx)
	if err != nil {
		return fmt.Errorf("get SWO shared session lock: %w", err)
	}

	state, err := swodb.New(c.Conn()).CurrentSwitchoverState(ctx)
	if err != nil {
		return fmt.Errorf("get current SWO state: %w", err)
	}

	if state == swodb.EnumSwitchoverStateUseNextDb {
		return swogrp.ErrDone
	}

	return nil
}
