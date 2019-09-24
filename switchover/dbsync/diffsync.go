package dbsync

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

const batchSize = 100

func (s *Sync) diffSync(ctx context.Context, txSrc, txDst pgx.Tx, dstChange int) error {
	start := time.Now()
	rows, err := txSrc.Query(ctx, `
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
	`, dstChange)
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
		case "INIT":
			continue
		case "DELETE":
			query = s.table(c.Table).DeleteOneRow()
		case "INSERT":
			query = s.table(c.Table).InsertOneRow()
		case "UPDATE":
			query = s.table(c.Table).UpdateOneRow()
		}

		_, err = txDst.Prepare(ctx, name, query)
		if err != nil {
			return errors.Wrap(err, "prepare statement")
		}
	}

	_, err = txDst.Prepare(ctx, "_ins:change_log", `
		insert into change_log (id, op, table_name, row_id)
		values ($1, $2, $3, $4)
	`)
	if err != nil {
		return errors.Wrap(err, "prepare statement")
	}
	fmt.Println("Prepared statements in", time.Since(start))

	p := mpb.NewWithContext(ctx)
	bar := p.AddBar(int64(len(changes)),
		mpb.BarClearOnComplete(),
		mpb.PrependDecorators(
			decor.CountersNoUnit("Synced %d of %d changes"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Elapsed(decor.ET_STYLE_GO), ""),
		),
	)

	var batchCount int
	b := &pgx.Batch{}
	for _, c := range changes {
		switch c.OP {
		case "DELETE":
			b.Queue(c.OP+":"+c.Table, c.RowID)
		case "INSERT":
			b.Queue(c.OP+":"+c.Table, c.RowData)
		case "UPDATE":
			b.Queue(c.OP+":"+c.Table, c.RowID, c.RowData)
		}
		b.Queue("_ins:change_log", c.ID, c.OP, c.Table, c.RowID)
		batchCount++
		if batchCount >= batchSize {
			err = txDst.SendBatch(ctx, b).Close()
			if err != nil {
				bar.Abort(false)
				p.Wait()
				return err
			}
			bar.IncrBy(batchCount)
			b = &pgx.Batch{}
			batchCount = 0
		}
	}
	if batchCount > 0 {
		err = txDst.SendBatch(ctx, b).Close()
		if err != nil {
			bar.Abort(false)
			p.Wait()
			return errors.Wrap(err, "sync")
		}
		bar.IncrBy(batchCount)
	}

	p.Wait()
	return nil
}
