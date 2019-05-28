package dbsync

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func (s *Sync) initialSync(ctx context.Context, txSrc, txDst *pgx.Tx) error {
	p := mpb.New()
	var err error
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
		err := txSrc.QueryRowEx(ctx, `select count(*) from `+t.SafeName(), nil).Scan(&rowCount)
		if err != nil {
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
			p.Abort(bars[i], false)
		}
		p.Abort(tBar, false)
		p.Wait()
	}

	for i, t := range toSync {
		err = func() error {
			defer tBar.Increment()

			pr, pw := io.Pipe()
			bw := bufio.NewWriter(pw)
			br := bufio.NewReader(pr)
			errCh := make(chan error, 2)
			go func() {
				defer pw.Close()
				defer bw.Flush()
				errCh <- errors.Wrap(txSrc.CopyToWriter(pw, fmt.Sprintf(`copy %s to stdout`, t.SafeName())), "read from src")
			}()
			go func() {
				r := io.TeeReader(br, &progWrite{inc1: tBar.IncrBy, inc2: bars[i].IncrBy})
				errCh <- errors.Wrap(txDst.CopyFromReader(r, fmt.Sprintf(`copy %s from stdin`, t.SafeName())), "write to dst")
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
