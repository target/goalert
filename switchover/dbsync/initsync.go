package dbsync

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/sqlutil"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func (s *Sync) initialSync(ctx context.Context, txSrc, txDst pgx.Tx) error {
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
		err := txSrc.QueryRow(ctx, `select count(*) from `+t.SafeName()).Scan(&rowCount)
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
			safeCols := t.ColumnNames()
			for i, n := range safeCols {
				safeCols[i] = sqlutil.QuoteID(n)
			}
			rows, err := txSrc.Query(ctx, "select "+strings.Join(safeCols, ",")+" from "+t.SafeName())
			if err != nil {
				return err
			}
			defer rows.Close()

			_, err = txDst.CopyFrom(ctx, pgx.Identifier{t.Name}, t.ColumnNames(), &progWrite{rows: rows, inc1: tBar.IncrBy, inc2: bars[i].IncrBy})
			return err
		}()
		if err != nil {
			abort(i)
			return err
		}
	}

	p.Wait()
	return nil
}
