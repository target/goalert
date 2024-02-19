// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package swodb

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const activeTxCount = `-- name: ActiveTxCount :one
SELECT COUNT(*)
FROM pg_stat_activity
WHERE "state" <> 'idle'
    AND "xact_start" <= $1
`

func (q *Queries) ActiveTxCount(ctx context.Context, xactStart pgtype.Timestamptz) (int64, error) {
	row := q.db.QueryRow(ctx, activeTxCount, xactStart)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const connectionInfo = `-- name: ConnectionInfo :many
SELECT application_name AS NAME,
    COUNT(*)
FROM pg_stat_activity
WHERE datname = current_database()
GROUP BY NAME
`

type ConnectionInfoRow struct {
	Name  pgtype.Text
	Count int64
}

func (q *Queries) ConnectionInfo(ctx context.Context) ([]ConnectionInfoRow, error) {
	rows, err := q.db.Query(ctx, connectionInfo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ConnectionInfoRow
	for rows.Next() {
		var i ConnectionInfoRow
		if err := rows.Scan(&i.Name, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const databaseInfo = `-- name: DatabaseInfo :one
SELECT db_id AS id,
    version()
FROM switchover_state
`

type DatabaseInfoRow struct {
	ID      pgtype.UUID
	Version string
}

func (q *Queries) DatabaseInfo(ctx context.Context) (DatabaseInfoRow, error) {
	row := q.db.QueryRow(ctx, databaseInfo)
	var i DatabaseInfoRow
	err := row.Scan(&i.ID, &i.Version)
	return i, err
}

const disableChangeLogTriggers = `-- name: DisableChangeLogTriggers :exec
UPDATE switchover_state
SET current_state = 'idle'
WHERE current_state = 'in_progress'
`

func (q *Queries) DisableChangeLogTriggers(ctx context.Context) error {
	_, err := q.db.Exec(ctx, disableChangeLogTriggers)
	return err
}

const enableChangeLogTriggers = `-- name: EnableChangeLogTriggers :exec
UPDATE switchover_state
SET current_state = 'in_progress'
WHERE current_state = 'idle'
`

func (q *Queries) EnableChangeLogTriggers(ctx context.Context) error {
	_, err := q.db.Exec(ctx, enableChangeLogTriggers)
	return err
}

const foreignKeyRefs = `-- name: ForeignKeyRefs :many
SELECT src.relname::text,
    dst.relname::text
FROM pg_catalog.pg_constraint con
    JOIN pg_catalog.pg_namespace ns ON ns.nspname = 'public'
    AND ns.oid = con.connamespace
    JOIN pg_catalog.pg_class src ON src.oid = con.conrelid
    JOIN pg_catalog.pg_class dst ON dst.oid = con.confrelid
WHERE con.contype = 'f'
    AND NOT con.condeferrable
`

type ForeignKeyRefsRow struct {
	SrcRelname string
	DstRelname string
}

func (q *Queries) ForeignKeyRefs(ctx context.Context) ([]ForeignKeyRefsRow, error) {
	rows, err := q.db.Query(ctx, foreignKeyRefs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ForeignKeyRefsRow
	for rows.Next() {
		var i ForeignKeyRefsRow
		if err := rows.Scan(&i.SrcRelname, &i.DstRelname); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const lastLogID = `-- name: LastLogID :one
SELECT COALESCE(MAX(id), 0)::bigint
FROM switchover_log
`

func (q *Queries) LastLogID(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, lastLogID)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}

const logEvents = `-- name: LogEvents :many
SELECT id,
    TIMESTAMP,
    DATA
FROM switchover_log
WHERE id > $1
ORDER BY id ASC
LIMIT 100
`

func (q *Queries) LogEvents(ctx context.Context, id int64) ([]SwitchoverLog, error) {
	rows, err := q.db.Query(ctx, logEvents, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SwitchoverLog
	for rows.Next() {
		var i SwitchoverLog
		if err := rows.Scan(&i.ID, &i.Timestamp, &i.Data); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const now = `-- name: Now :one
SELECT now()::timestamptz
`

func (q *Queries) Now(ctx context.Context) (pgtype.Timestamptz, error) {
	row := q.db.QueryRow(ctx, now)
	var column_1 pgtype.Timestamptz
	err := row.Scan(&column_1)
	return column_1, err
}

const sequenceNames = `-- name: SequenceNames :many
SELECT sequence_name::text
FROM information_schema.sequences
WHERE sequence_catalog = current_database()
    AND sequence_schema = 'public'
    AND sequence_name != 'change_log_id_seq'
`

func (q *Queries) SequenceNames(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, sequenceNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var sequence_name string
		if err := rows.Scan(&sequence_name); err != nil {
			return nil, err
		}
		items = append(items, sequence_name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const tableColumns = `-- name: TableColumns :many
SELECT col.table_name::text,
    col.column_name::text,
    col.data_type::text,
    col.ordinal_position::INT
FROM information_schema.columns col
    JOIN information_schema.tables t ON t.table_catalog = col.table_catalog
    AND t.table_schema = col.table_schema
    AND t.table_name = col.table_name
    AND t.table_type = 'BASE TABLE'
WHERE col.table_catalog = current_database()
    AND col.table_schema = 'public'
`

type TableColumnsRow struct {
	ColTableName       string
	ColColumnName      string
	ColDataType        string
	ColOrdinalPosition int32
}

func (q *Queries) TableColumns(ctx context.Context) ([]TableColumnsRow, error) {
	rows, err := q.db.Query(ctx, tableColumns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TableColumnsRow
	for rows.Next() {
		var i TableColumnsRow
		if err := rows.Scan(
			&i.ColTableName,
			&i.ColColumnName,
			&i.ColDataType,
			&i.ColOrdinalPosition,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
