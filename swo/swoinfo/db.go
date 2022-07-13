package swoinfo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swodb"
)

type DB struct {
	ID      uuid.UUID
	Version string
}

func DBInfo(ctx context.Context, conn *pgx.Conn) (*DB, error) {
	info, err := swodb.New(conn).DatabaseInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &DB{
		ID:      info.ID,
		Version: info.Version,
	}, nil
}
