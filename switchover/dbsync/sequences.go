package dbsync

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/target/goalert/util/sqlutil"
)

func (s *Sync) syncSequences(ctx context.Context, txSrc, txDst pgx.Tx) error {
	rows, err := txSrc.Query(ctx, `
		select sequence_name
		from information_schema.sequences
		where
			sequence_catalog = current_database() and
			sequence_schema = 'public'
	`)
	if err != nil {
		return errors.Wrap(err, "get sequence names")
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return errors.Wrap(err, "scan sequence name")
		}
		names = append(names, name)
	}
	rows.Close()
	batchRead := &pgx.Batch{}
	for _, name := range names {
		batchRead.Queue(`select last_value, is_called from ` + sqlutil.QuoteID(name))
	}
	readResults := txSrc.SendBatch(ctx, batchRead)
	if err != nil {
		return errors.Wrap(err, "send src sequence queries")
	}
	batchWrite := &pgx.Batch{}
	for _, name := range names {
		var lastVal int64
		var called bool
		err = readResults.QueryRow().Scan(&lastVal, &called)
		if err != nil {
			return errors.Wrapf(err, "get src sequence state for '%s'", name)
		}
		batchWrite.Queue(`select pg_catalog.setval($1, $2, $3)`, name, lastVal, called)
	}
	err = readResults.Close()
	if err != nil {
		return errors.Wrap(err, "close src data")
	}

	writeResults := txDst.SendBatch(ctx, batchWrite)
	return errors.Wrap(writeResults.Close(), "update dst sequence state")
}
