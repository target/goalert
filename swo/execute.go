package swo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/swo/swosync"
)

func (m *Manager) DoExecute(ctx context.Context) error {
	return m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		rep := swosync.NewLogicalReplicator()
		rep.SetSourceDB(oldConn)
		rep.SetDestinationDB(newConn)
		rep.SetProgressFunc(swogrp.Progressf)

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

		err = m.PauseApps(ctx)
		if err != nil {
			return fmt.Errorf("pause apps: %w", err)
		}

		for i := 0; i < 10; i++ {
			err = rep.LogicalSync(ctx)
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
}
