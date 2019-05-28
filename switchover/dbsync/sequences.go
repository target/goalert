package dbsync

import (
	"context"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
)

func (s *Sync) syncSequences(ctx context.Context, txSrc, txDst *pgx.Tx) error {
	rows, err := txSrc.QueryEx(ctx, `
		select sequence_name
		from information_schema.sequences
		where
			sequence_catalog = current_database() and
			sequence_schema = 'public'
	`, nil)
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
	batchRead := txSrc.BeginBatch()
	for _, name := range names {
		batchRead.Queue(`select last_value, is_called from `+pgx.Identifier{name}.Sanitize(), nil, nil, []int16{pgx.BinaryFormatCode, pgx.BinaryFormatCode})
	}
	err = batchRead.Send(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "send src sequence queries")
	}
	batch := txDst.BeginBatch()
	for _, name := range names {
		var lastVal int64
		var called bool
		err = batchRead.QueryRowResults().Scan(&lastVal, &called)
		if err != nil {
			return errors.Wrapf(err, "get src sequence state for '%s'", name)
		}
		batch.Queue(`select pg_catalog.setval($1, $2, $3)`, []interface{}{name, lastVal, called}, []pgtype.OID{pgtype.TextOID, pgtype.Int8OID, pgtype.BoolOID}, nil)
	}
	err = batch.Send(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "update dst sequence state")
	}

	return errors.Wrap(batch.Close(), "update dst sequence state")
}
