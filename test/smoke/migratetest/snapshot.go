package migratetest

import (
	"bytes"
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/devtools/pgdump-lite"
)

// Snapshot is a snapshot of a database's schema and data.
type Snapshot struct {
	Schema    *pgdump.Schema
	TableData []TableSnapshot
}

var snapshotBuf bytes.Buffer
var mx sync.Mutex

// NewSnapshotURL will create a new Snapshot from a database URL.
func NewSnapshotURL(ctx context.Context, dbURL string) (*Snapshot, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewSnapshot(ctx, db)
}

// NewSnapshot will create a new Snapshot from a database connection.
func NewSnapshot(ctx context.Context, db *pgxpool.Pool) (*Snapshot, error) {
	mx.Lock()
	defer mx.Unlock()

	schema, err := pgdump.DumpSchema(ctx, db)
	if err != nil {
		return nil, err
	}

	snapshotBuf.Reset()
	err = pgdump.DumpDataWithSchemaParallel(ctx, db, &snapshotBuf, nil, schema)
	if err != nil {
		return nil, err
	}

	scan := NewCopyScanner(&snapshotBuf)
	var tables []TableSnapshot
	for scan.Scan() {
		tables = append(tables, scan.Table())
	}
	if scan.Err() != nil {
		return nil, scan.Err()
	}

	return &Snapshot{
		Schema:    schema,
		TableData: tables,
	}, nil
}
