package swo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/util/sqldrv"
)

// NewAppPGXPool returns a pgxpool.Pool that will use the old database until the
// switchover_state table indicates that the new database should be used.
//
// Until the switchover is complete, the old database will be protected with a
// shared advisory lock (4369).
func NewAppPGXPool(oldURL, nextURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(oldURL)
	if err != nil {
		return nil, fmt.Errorf("parse old URL: %w", err)
	}
	sqldrv.SetConfigRetries(cfg)
	nextCfg, err := pgxpool.ParseConfig(nextURL)
	if err != nil {
		return nil, fmt.Errorf("parse next URL: %w", err)
	}
	sqldrv.SetConfigRetries(nextCfg)

	// speed up cleanup
	cfg.HealthCheckPeriod = 100 * time.Millisecond
	cfg.MaxConnIdleTime = time.Second

	var mx sync.Mutex
	var isDone bool

	cfg.BeforeConnect = func(ctx context.Context, cfg *pgx.ConnConfig) error {
		mx.Lock()
		defer mx.Unlock()

		if isDone {
			// switched, use new config
			*cfg = *nextCfg.ConnConfig
		}
		return nil
	}

	cfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		useNext, err := swodb.New(conn).SWOConnLock(ctx)
		if err != nil {
			// error, don't use
			return false
		}
		if !useNext {
			return true
		}

		// switch to new config, don't use this connection
		mx.Lock()
		isDone = true
		mx.Unlock()

		return false
	}

	cfg.AfterRelease = func(conn *pgx.Conn) bool {
		err := swodb.New(conn).SWOConnUnlockAll(context.Background())
		return err == nil
	}

	return pgxpool.NewWithConfig(context.Background(), cfg)
}
