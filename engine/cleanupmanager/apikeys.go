package cleanupmanager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
)

type APIKeysArgs struct{}

func (APIKeysArgs) Kind() string { return "cleanup-manager-api-keys" }

// CleanupAPIKeys will revoke access to the API from unused tokens, including both user sessions and calendar subscriptions.
func (db *DB) CleanupAPIKeys(ctx context.Context, j *river.Job[APIKeysArgs]) error {
	err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		// After 30 days, the token is no longer valid, so delete it.
		//
		// This is defined by how the keyring system works for session signing, and is not influenced by the APIKeyExpireDays config.
		count, err := gadb.New(tx).CleanupMgrDeleteOldSessions(ctx, 30)
		if err != nil {
			return false, fmt.Errorf("delete old user sessions: %w", err)
		}
		return count < 100, nil
	})
	if err != nil {
		return err
	}

	cfg := config.FromContext(ctx)
	if cfg.Maintenance.APIKeyExpireDays <= 0 {
		return nil
	}

	err = db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		count, err := gadb.New(tx).CleanupMgrDisableOldCalSub(ctx, int32(cfg.Maintenance.APIKeyExpireDays))
		if err != nil {
			return false, fmt.Errorf("disable unused calsub keys: %w", err)
		}
		return count < 100, nil
	})
	if err != nil {
		return err
	}

	return nil
}
