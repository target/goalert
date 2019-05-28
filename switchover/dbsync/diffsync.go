package dbsync

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

const batchSize = 100

type decorElapsedDone time.Time

func (d decorElapsedDone) Decor(stats *decor.Statistics) string {
	if !stats.Completed {
		return ""
	}
	return " in " + time.Since(time.Time(d)).String()
}
func (d decorElapsedDone) Syncable() (bool, chan int) {
	return false, nil
}

func (s *Sync) diffSync(ctx context.Context, txSrc, txDst *pgx.Tx, dstChange int) error {
	start := time.Now()
	rows, err := txSrc.QueryEx(ctx, `
		with tx_max_id as (
			select max(id), tx_id
			from change_log
			where id > $1
			group by tx_id
		)
		select id, op, table_name, row_id, row_data
		from change_log c
		join tx_max_id max_id on max_id.tx_id = c.tx_id
		where c.id > $1
		order by
			max_id.max,
			cmd_id::text::int
	`, nil, dstChange)
	if err != nil {
		return errors.Wrap(err, "get changed rows")
	}
	defer rows.Close()
	type change struct {
		ID      int
		OP      string
		Table   string
		RowID   string
		RowData []byte
	}

	var changes []change
	for rows.Next() {
		var c change
		err = rows.Scan(&c.ID, &c.OP, &c.Table, &c.RowID, &c.RowData)
		if err != nil {
			return errors.Wrap(err, "read change")
		}
		changes = append(changes, c)
	}
	rows.Close()
	fmt.Printf("Fetched %d changes in %s\n", len(changes), time.Since(start))

	start = time.Now()
	// prepare statements
	for _, c := range changes {
		name := c.OP + ":" + c.Table
		var query string
		switch c.OP {
		case "DELETE":
			query = s.table(c.Table).DeleteOneRow()
		case "INSERT":
			query = s.table(c.Table).InsertOneRow()
		case "UPDATE":
			query = s.table(c.Table).UpdateOneRow()
		}
		_, err = txDst.PrepareEx(ctx, name, query, nil)
		if err != nil {
			return errors.Wrap(err, "prepare statement")
		}
	}

	_, err = txDst.PrepareEx(ctx, "_ins:change_log", `
		insert into change_log (id, op, table_name, row_id)
		values ($1, $2, $3, $4)
	`, nil)
	if err != nil {
		return errors.Wrap(err, "prepare statement")
	}
	fmt.Println("Prepared statements in", time.Since(start))

	p := mpb.New()
	bar := p.AddBar(int64(len(changes)),
		mpb.BarClearOnComplete(),
		mpb.PrependDecorators(
			decor.CountersNoUnit("Synced %d of %d changes"),
		),
		mpb.AppendDecorators(
			decorElapsedDone(time.Now()),
		),
	)

	var batchCount int
	b := txDst.BeginBatch()
	for _, c := range changes {
		switch c.OP {
		case "DELETE":
			b.Queue(c.OP+":"+c.Table, []interface{}{c.RowID}, nil, nil)
		case "INSERT":
			b.Queue(c.OP+":"+c.Table, []interface{}{c.RowData}, nil, nil)
		case "UPDATE":
			b.Queue(c.OP+":"+c.Table, []interface{}{c.RowID, c.RowData}, nil, nil)
		}
		b.Queue("_ins:change_log", []interface{}{c.ID, c.OP, c.Table, c.RowID}, nil, nil)
		batchCount++
		if batchCount >= batchSize {
			err = b.Send(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "send batched commands")
			}
			err = b.Close()
			if err != nil {
				p.Abort(bar, false)
				p.Wait()
				fmt.Println("SYNC", c.ID)
				return err
			}
			bar.IncrBy(batchCount)
			b = txDst.BeginBatch()
			batchCount = 0
		}
	}
	if batchCount > 0 {
		err = b.Send(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "sync")
		}
		err = b.Close()
		if err != nil {
			p.Abort(bar, false)
			p.Wait()
			return errors.Wrap(err, "sync")
		}
		bar.IncrBy(batchCount)
	}

	p.Wait()
	return nil
}
