package swo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/sqlutil"
)

func (e *Execute) syncSequences(ctx context.Context, src, dst pgx.Tx) error {
	go e.Progressf(ctx, "syncing sequences")
	var seqRead pgx.Batch
	for _, name := range e.seqNames {
		seqRead.Queue("select last_value, is_called from " + sqlutil.QuoteID(name))
	}

	res := src.SendBatch(ctx, &seqRead)
	var setSeq pgx.Batch
	for _, name := range e.seqNames {
		var last int64
		var called bool
		err := res.QueryRow().Scan(&last, &called)
		if err != nil {
			return fmt.Errorf("get sequence %s: %w", name, err)
		}
		setSeq.Queue("select pg_catalog.setval($1, $2, $3)", name, last, called)
	}
	if err := res.Close(); err != nil {
		return fmt.Errorf("close seq batch: %w", err)
	}

	err := dst.SendBatch(ctx, &setSeq).Close()
	if err != nil {
		return fmt.Errorf("set sequences: %w", err)
	}

	return nil
}
