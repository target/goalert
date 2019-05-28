package heartbeatmanager

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// UpdateAll will process all heartbeats opening and closing alerts as needed.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := db.processAll(ctx)
	return err
}

func (db *DB) processAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Processing heartbeats.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer tx.Rollback()

	var newAlertCtx []context.Context
	var newAlerts []alert.Alert
	bad, err := db.unhealthy(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "fetch unhealthy heartbeats")
	}
	for _, row := range bad {
		a, isNew, err := db.alertStore.CreateOrUpdateTx(row.Context(ctx), tx, &alert.Alert{
			Summary:   fmt.Sprintf("Heartbeat monitor '%s' expired.", row.Name),
			Details:   "Last heartbeat: " + row.LastHeartbeat.Format(time.UnixDate),
			Status:    alert.StatusTriggered,
			ServiceID: row.ServiceID,
			Dedup: &alert.DedupID{
				Type:    alert.DedupTypeHeartbeat,
				Version: 1,
				Payload: row.ID,
			},
		})
		if err != nil {
			return errors.Wrap(err, "create alert")
		}
		if isNew {
			// Store contexts with alert info for each alert that was newly-created.
			newAlertCtx = append(newAlertCtx, log.WithFields(row.Context(ctx), log.Fields{
				"AlertID":   a.ID,
				"ServiceID": a.ServiceID,
			}))
			newAlerts = append(newAlerts, *a)
		}
	}
	good, err := db.healthy(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "fetch healthy heartbeats")
	}
	for _, row := range good {
		_, _, err = db.alertStore.CreateOrUpdateTx(row.Context(ctx), tx, &alert.Alert{
			Status:    alert.StatusClosed,
			ServiceID: row.ServiceID,
			Dedup: &alert.DedupID{
				Type:    alert.DedupTypeHeartbeat,
				Version: 1,
				Payload: row.ID,
			},
		})
		if err != nil {
			return errors.Wrap(err, "close alert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// log new alert creations, after the tx was committed without err.
	for _, ctx := range newAlertCtx {
		log.Logf(ctx, "Alert created.")

	}
	for _, n := range newAlerts {
		trace.FromContext(ctx).Annotate(
			[]trace.Attribute{
				trace.StringAttribute("service.id", n.ServiceID),
				trace.Int64Attribute("alert.id", int64(n.ID)),
			},
			"Alert created.",
		)
	}
	return nil
}

type row struct {
	ID            string
	Name          string
	ServiceID     string
	LastHeartbeat time.Time
}

func (r row) Context(ctx context.Context) context.Context {
	return permission.ServiceSourceContext(permission.WithoutAuth(ctx), r.ServiceID, &permission.SourceInfo{
		Type: permission.SourceTypeHeartbeat,
		ID:   r.ID,
	})
}

func (db *DB) unhealthy(ctx context.Context, tx *sql.Tx) ([]row, error) {
	rows, err := tx.Stmt(db.fetchFailed).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []row
	for rows.Next() {
		var r row
		err = rows.Scan(&r.ID, &r.Name, &r.ServiceID, &r.LastHeartbeat)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
func (db *DB) healthy(ctx context.Context, tx *sql.Tx) ([]row, error) {
	rows, err := tx.Stmt(db.fetchHealthy).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []row
	for rows.Next() {
		var r row
		err = rows.Scan(&r.ID, &r.ServiceID)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
