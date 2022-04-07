package swo

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/swo/swogrp"
)

// WaitForActiveTx waits for all currently active transactions to complete in the main DB.
func (e *Execute) WaitForActiveTx(ctx context.Context) {
	if e.err != nil {
		return
	}

	swogrp.Progressf(ctx, "waiting for in-flight transactions to finish")

	var now time.Time
	err := e.mainDBConn.QueryRow(ctx, "select now()").Scan(&now)
	if err != nil {
		e.err = fmt.Errorf("wait for active tx: get current time: %w", err)
		return
	}

	for {
		var n int
		err = e.mainDBConn.QueryRow(ctx, "select count(*) from pg_stat_activity where state <> 'idle' and xact_start <= $1", now).Scan(&n)
		if err != nil {
			e.err = fmt.Errorf("wait for active tx: get active tx count: %w", err)
			return
		}
		if n == 0 {
			break
		}

		swogrp.Progressf(ctx, "waiting for %d transaction(s) to finish", n)
		time.Sleep(time.Second)
	}
}
