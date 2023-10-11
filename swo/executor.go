package swo

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/swo/swosync"
)

// Executor is responsible for executing the switchover process.
type Executor struct {
	mgr *Manager

	stateCh chan execState

	wf  *WithFunc[*swosync.LogicalReplicator]
	rep *swosync.LogicalReplicator
	mx  sync.Mutex
}

// NewExecutor initializes a new Executor for the given Manager.
func NewExecutor(mgr *Manager) *Executor {
	e := &Executor{
		mgr:     mgr,
		stateCh: make(chan execState, 1),
	}
	e.stateCh <- execStateIdle
	e.wf = NewWithFunc(func(ctx context.Context, fn func(*swosync.LogicalReplicator)) error {
		return mgr.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
			rep := swosync.NewLogicalReplicator()
			rep.SetSourceDB(oldConn)
			rep.SetDestinationDB(newConn)
			rep.SetProgressFunc(mgr.taskMgr.Statusf)
			fn(rep)
			return nil
		})
	})
	return e
}

type execState int

const (
	execStateIdle execState = iota
	execStateSync
)

var _ swogrp.Executor = (*Executor)(nil)

// Sync begins the switchover process by resetting and starting the logical replication process.
func (e *Executor) Sync(ctx context.Context) error {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.rep != nil {
		return fmt.Errorf("already syncing")
	}

	rep, err := e.wf.Begin(e.mgr.Logger.BackgroundContext())
	if err != nil {
		return err
	}

	err = rep.ResetChangeTracking(ctx)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	err = rep.StartTrackingChanges(ctx)
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	err = rep.FullInitialSync(ctx)
	if err != nil {
		return fmt.Errorf("initial sync: %w", err)
	}

	for i := 0; i < 10; i++ {
		err = rep.LogicalSync(ctx)
		if err != nil {
			return fmt.Errorf("logical sync: %w", err)
		}
	}

	e.rep = rep
	return nil
}

// Exec executes the switchover process, blocking until it is complete.
func (e *Executor) Exec(ctx context.Context) error {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.rep == nil {
		return fmt.Errorf("not syncing")
	}

	rep := e.rep
	e.rep = nil

	for i := 0; i < 10; i++ {
		err := rep.LogicalSync(ctx)
		if err != nil {
			return fmt.Errorf("logical sync (after pause): %w", err)
		}
	}

	err := rep.FinalSync(ctx)
	if err != nil {
		return fmt.Errorf("final sync: %w", err)
	}

	return nil
}

// Cancel cancels the switchover process.
func (e *Executor) Cancel() { e.wf.Cancel() }
