package swo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
)

func (m *Manager) Progressf(ctx context.Context, format string, a ...interface{}) {
}

func (m *Manager) InitialSync(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	m.Progressf(ctx, "scanning tables")
	tables, err := ScanTables(ctx, oldConn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	srcTx, err := oldConn.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
		IsoLevel:   pgx.Serializable,
	})
	if err != nil {
		return fmt.Errorf("begin src tx: %w", err)
	}
	defer srcTx.Rollback(ctx)

	dstTx, err := newConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin dst tx: %w", err)
	}
	defer dstTx.Rollback(ctx)

	for _, table := range tables {
		if table.SkipSync() {
			continue
		}

		err = m.SyncTableInit(ctx, table, srcTx, dstTx)
		if err != nil {
			return fmt.Errorf("sync table %s: %w", table.Name, err)
		}
	}

	m.Progressf(ctx, "commit initial sync")
	// Important to validate src commit, even though it's read-only.
	//
	// A failure here indicates the isolation level has been violated
	// and we will need to try again.
	err = srcTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit src tx: %w", err)
	}

	err = dstTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit dst tx: %w", err)
	}

	return nil
}

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

func (m *Manager) SyncTableInit(ctx context.Context, t Table, srcTx, dstTx pgx.Tx) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var rowCount int
	err := srcTx.QueryRow(ctx, fmt.Sprintf("select count(*) from %s", t.QuotedName())).Scan(&rowCount)
	if err != nil {
		return fmt.Errorf("count rows: %w", err)
	}

	pr, pw := io.Pipe()
	var lc lineCount
	errCh := make(chan error, 3)
	go func() {
		prog := time.NewTimer(2 * time.Second)
		defer prog.Stop()
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				pw.CloseWithError(ctx.Err())
				pr.CloseWithError(ctx.Err())
				return
			case <-prog.C:
			}

			m.Progressf(ctx, "syncing table %s (%d/%d)", t.Name, lc.Lines(), rowCount)
		}
	}()
	go func() {
		defer cancel()
		_, err := srcTx.Conn().PgConn().CopyTo(ctx, pw, fmt.Sprintf(`copy %s to stdout`, t.QuotedName()))
		if err != nil {
			errCh <- fmt.Errorf("read from src: %w", err)
			pw.CloseWithError(err)
			pr.CloseWithError(err)
		} else {
			errCh <- nil
		}
	}()
	go func() {
		defer cancel()
		_, err := dstTx.Conn().PgConn().CopyFrom(ctx, io.TeeReader(pr, &lc), fmt.Sprintf(`copy %s from stdin`, t.QuotedName()))
		if err != nil {
			errCh <- fmt.Errorf("write to dst: %w", err)
			pw.CloseWithError(err)
			pr.CloseWithError(err)
		} else {
			errCh <- nil
		}
	}()

	// check first error, but wait for all to finish
	err = <-errCh
	<-errCh
	<-errCh
	if err != nil {
		return err
	}

	return nil
}
