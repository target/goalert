package swo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swogrp"
)

type lineCount struct {
	n  int
	mx sync.Mutex
}

func (lc *lineCount) Write(p []byte) (n int, err error) {
	lc.mx.Lock()
	lc.n += bytes.Count(p, []byte("\n"))
	lc.mx.Unlock()
	return len(p), nil
}

func (lc *lineCount) Lines() int {
	lc.mx.Lock()
	defer lc.mx.Unlock()
	return lc.n
}

func (e *Execute) Progressf(ctx context.Context, format string, args ...interface{}) {
	if e.err != nil {
		return
	}

	swogrp.Progressf(ctx, format, args...)
}

func (e *Execute) do(ctx context.Context, desc string, fn func(context.Context) error) {
	if e.err != nil {
		return
	}

	e.err = fn(ctx)
	if e.err != nil {
		e.err = fmt.Errorf("%s: %w", desc, e.err)
	}
}

// SyncFull performs a full initial sync of the database by copying contents of each table directly to the
// destination database.
func (e *Execute) SyncFull(ctx context.Context) {
	if e.err != nil {
		return
	}
	e.Progressf(ctx, "performing initial sync")

	srcTx, dstTx, err := e.syncTx(ctx, true)
	if err != nil {
		e.err = fmt.Errorf("initial sync: begin: %w", err)
		return
	}
	defer srcTx.Rollback(ctx)
	defer dstTx.Rollback(ctx)

	// defer all constraints
	if _, err = dstTx.Exec(ctx, "SET CONSTRAINTS ALL DEFERRED"); err != nil {
		e.err = fmt.Errorf("initial sync: defer constraints: %w", err)
		return
	}

	for _, table := range e.tables {
		if table.SkipSync() {
			continue
		}

		if err = e.syncTableFull(ctx, table, srcTx, dstTx); err != nil {
			e.err = fmt.Errorf("initial sync: copy %s: %w", table.Name, err)
			return
		}
	}

	e.Progressf(ctx, "commit initial sync")
	// Important to validate src commit, even though it's read-only.
	//
	// A failure here indicates the isolation level has been violated
	// and we will need to try again.
	if err := srcTx.Commit(ctx); err != nil {
		e.err = fmt.Errorf("initial sync: src commit: %w", err)
		return
	}
	if err := dstTx.Commit(ctx); err != nil {
		e.err = fmt.Errorf("initial sync: dst commit: %w", err)
		return
	}

	// vacuum analyze new DB
	e.Progressf(ctx, "vacuum analyze")
	if _, err := e.nextDBConn.Exec(ctx, "VACUUM ANALYZE"); err != nil {
		e.err = fmt.Errorf("initial sync: vacuum analyze: %w", err)
		return
	}
}

// syncTableFull will copy the contents of the table from the source database to the destination database using
// COPY TO and COPY FROM.
func (e *Execute) syncTableFull(origCtx context.Context, t Table, srcTx, dstTx pgx.Tx) error {
	ctx, cancel := context.WithCancel(origCtx)
	defer cancel()

	var rowCount int
	err := srcTx.QueryRow(ctx, fmt.Sprintf("select count(*) from %s", t.QuotedName())).Scan(&rowCount)
	if err != nil {
		return fmt.Errorf("sync table %s: get row count: %w", t.Name, err)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	pr, pw := io.Pipe()
	var lc lineCount
	go func() {
		defer wg.Done()
		prog := time.NewTimer(500 * time.Millisecond)
		defer prog.Stop()
		for {
			swogrp.Progressf(origCtx, "syncing table %s (%d/%d)", t.Name, lc.Lines(), rowCount)
			select {
			case <-ctx.Done():
				pw.CloseWithError(ctx.Err())
				pr.CloseWithError(ctx.Err())
				return
			case <-prog.C:
			}
		}
	}()

	var srcErr, dstErr error

	go func() {
		defer wg.Done()
		_, srcErr = srcTx.Conn().PgConn().CopyTo(ctx, pw, fmt.Sprintf(`copy %s to stdout`, t.QuotedName()))
		pw.Close()
	}()
	go func() {
		defer wg.Done()
		defer cancel()
		_, dstErr = dstTx.Conn().PgConn().CopyFrom(ctx, io.TeeReader(pr, &lc), fmt.Sprintf(`copy %s from stdin`, t.QuotedName()))
		pr.Close()
	}()

	wg.Wait()

	if dstErr != nil {
		return fmt.Errorf("copy to dst: %w", dstErr)
	}
	if srcErr != nil {
		return fmt.Errorf("copy from src: %w", srcErr)
	}

	return nil
}
