package swo

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swogrp"
)

// WaitForActiveTx waits for all currently active transactions to complete in the main DB.
func (e *Execute) WaitForActiveTx(ctx context.Context) {
	if e.err != nil {
		return
	}

	swogrp.Progressf(ctx, "waiting for in-flight transactions to finish")

	db := swodb.New(e.mainDBConn)

	var now time.Time
	now, err := db.CurrentTime(ctx)
	if err != nil {
		e.err = fmt.Errorf("wait for active tx: get current time: %w", err)
		return
	}

	for {
		n, err := db.ActiveTxCount(ctx, now)
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
