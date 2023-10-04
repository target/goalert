package swoinfo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/swo/swodb"
)

// DB contains information about a database.
type DB struct {
	// ID is the UUID of the database, stored in the switchover_state table.
	ID      uuid.UUID
	Version string
}

// DBInfo provides information about the database associated with the given connection.
func DBInfo(ctx context.Context, conn *pgx.Conn) (*DB, error) {
	info, err := swodb.New(conn).DatabaseInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !info.ID.Valid {
		return nil, fmt.Errorf("no database ID")
	}
	return &DB{
		ID:      info.ID.Bytes,
		Version: info.Version,
	}, nil
}
