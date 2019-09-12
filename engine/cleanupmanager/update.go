package cleanupmanager

import (
	"context"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"net/http"
	"net/url"
)

// UpdateAll will update the state of all active escalation policies.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := db.update(ctx)
	return err
}

func (db *DB) update(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running cleanup operations.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.StmtContext(ctx, db.orphanSlackChan).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	var toDelete sqlutil.UUIDArray
	for rows.Next() {
		var id, token string
		err = rows.Scan(&id, &token)
		if err != nil {
			return err
		}
		log.Debugf(ctx, "cleanup notification channel %s", id)
		data, _, err := db.keys.Decrypt([]byte(token))
		if err != nil {
			return err
		}

		// TODO: implement retry/backoff logic?
		go http.Get("https://slack.com/api/auth.revoke?token=" + url.QueryEscape(string(data)))

		toDelete = append(toDelete, id)
	}

	_, err = tx.StmtContext(ctx, db.deleteChan).ExecContext(ctx, toDelete)
	if err != nil {
		return err
	}

	return tx.Commit()
}
