package swo

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/swo/swosync"
)

type Executor struct {
	mgr *Manager
	mx  sync.Mutex

	ctxCh  chan context.Context
	errCh  chan error
	cancel func()
}

var _ swogrp.Executor = (*Executor)(nil)

func (e *Executor) init() {
	e.mx.Lock()
	defer e.mx.Unlock()
	if e.cancel != nil {
		panic("already running")
	}

	ctx, cancel := context.WithCancel(e.mgr.Logger.BackgroundContext())
	e.cancel = cancel
	e.ctxCh = make(chan context.Context)
	e.errCh = make(chan error, 2)

	go func() {
		defer e.Cancel()
		e.errCh <- e.mgr.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
			rep := swosync.NewLogicalReplicator()
			rep.SetSourceDB(oldConn)
			rep.SetDestinationDB(newConn)
			rep.SetProgressFunc(e.mgr.taskMgr.Statusf)

			// sync
			ctx = <-e.ctxCh
			err := rep.Reset(ctx)
			if err != nil {
				return fmt.Errorf("reset: %w", err)
			}

			err = rep.Start(ctx)
			if err != nil {
				return fmt.Errorf("start: %w", err)
			}

			err = rep.InitialSync(ctx)
			if err != nil {
				return fmt.Errorf("initial sync: %w", err)
			}

			for i := 0; i < 10; i++ {
				err = rep.LogicalSync(ctx)
				if err != nil {
					return fmt.Errorf("logical sync: %w", err)
				}
			}
			e.errCh <- nil

			// wait for pause
			ctx = <-e.ctxCh
			for i := 0; i < 10; i++ {
				err := rep.LogicalSync(ctx)
				if err != nil {
					return fmt.Errorf("logical sync (after pause): %w", err)
				}
			}

			err = rep.FinalSync(ctx)
			if err != nil {
				return fmt.Errorf("final sync: %w", err)
			}

			return nil
		})
	}()
}

func (e *Executor) Sync(ctx context.Context) error {
	e.init()

	e.ctxCh <- ctx
	return <-e.errCh
}

func (e *Executor) Exec(ctx context.Context) error {
	e.ctxCh <- ctx
	return <-e.errCh
}

func (e *Executor) Cancel() {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.cancel == nil {
		return
	}

	e.cancel()
	e.ctxCh = nil
	e.errCh = nil
	e.cancel = nil
}
