package switchover

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

type dbState struct {
	timeOffset time.Duration
	dbc        driver.Connector
	db         *sql.DB
}

func newDBState(ctx context.Context, dbc driver.Connector) (*dbState, error) {
	db := sql.OpenDB(dbc)
	offset, err := CalcDBOffset(ctx, db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &dbState{
		dbc:        dbc,
		db:         db,
		timeOffset: offset,
	}, nil
}

func CalcDBOffset(ctx context.Context, db *sql.DB) (time.Duration, error) {
	s, err := db.PrepareContext(ctx, `select now()`)
	if err != nil {
		return 0, err
	}
	defer s.Close()

	// pre-run the query to reduce error of first run
	s.ExecContext(ctx)

	var sum time.Duration
	var t time.Time
	for i := 0; i < 10; i++ {
		err = s.QueryRowContext(ctx).Scan(&t)
		if err != nil {
			return 0, err
		}
		sum += time.Until(t)
	}
	return sum / 10, err
}
