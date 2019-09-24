package dbsync

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func (s *Sync) initialSync(ctx context.Context, src, dst *pgx.Conn) error {
	err := s.RefreshTables(ctx)
	if err != nil {
		return err
	}

	p := mpb.NewWithContext(ctx)
	var totalRows int64
	var bars []*mpb.Bar
	var toSync []Table
	scanBar := p.AddBar(int64(len(s.tables)),
		mpb.BarRemoveOnComplete(),
		mpb.BarPriority(9999),
		mpb.PrependDecorators(
			decor.CountersNoUnit("Scanning tables (%d of %d)...", decor.WCSyncSpaceR),
		),
	)
	for _, t := range s.tables {
		var rowCount int64
		err := src.QueryRow(ctx, `select count(*) from `+t.SafeName()).Scan(&rowCount)
		if err != nil {
			scanBar.Abort(false)
			p.Wait()
			return err
		}
		scanBar.Increment()
		if rowCount == 0 {
			continue
		}
		totalRows += rowCount
		bars = append(bars, p.AddBar(int64(rowCount),
			mpb.BarClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(t.Name, decor.WCSyncSpaceR),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(), "Done"),
			),
		))
		toSync = append(toSync, t)
	}
	tBar := p.AddBar(int64(totalRows),
		mpb.BarClearOnComplete(),
		mpb.PrependDecorators(
			decor.CountersNoUnit("Synced %d of %d rows", decor.WCSyncSpaceR),
		),
	)
	abort := func(i int) {
		for ; i < len(toSync); i++ {
			bars[i].Abort(false)
		}
		tBar.Abort(false)
		p.Wait()
	}

	for i, t := range toSync {
		err = func() error {
			defer tBar.Increment()

			pr, pw := io.Pipe()
			bw := bufio.NewWriter(pw)
			br := bufio.NewReader(pr)
			errCh := make(chan error, 3)
			go func() {
				<-ctx.Done()
				go pw.CloseWithError(ctx.Err())
				go pr.CloseWithError(ctx.Err())
				errCh <- ctx.Err()
			}()
			go func() {
				defer pw.Close()
				defer bw.Flush()
				_, err := src.PgConn().CopyTo(ctx, pw, fmt.Sprintf(`copy %s to stdout`, t.SafeName()))
				errCh <- errors.Wrap(err, "read from src")
			}()
			go func() {
				r := io.TeeReader(br, &progWrite{inc1: tBar.IncrBy, inc2: bars[i].IncrBy})
				_, err := dst.PgConn().CopyFrom(ctx, r, fmt.Sprintf(`copy %s from stdin`, t.SafeName()))
				errCh <- errors.Wrap(err, "write to dst")
			}()
			err = <-errCh
			if err != nil {
				return err
			}
			err = <-errCh
			if err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			abort(i)
			return err
		}
	}

	p.Wait()
	return nil
}
