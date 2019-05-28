package schedulemanager

import (
	"context"
	"database/sql"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = db.update(ctx)
	return err
}

func (db *DB) update(ctx context.Context) error {
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer tx.Rollback()
	log.Debugf(ctx, "Updating schedule rules.")

	var now time.Time
	err = tx.Stmt(db.currentTime).QueryRowContext(ctx).Scan(&now)
	if err != nil {
		return errors.Wrap(err, "get DB time")
	}

	rows, err := tx.Stmt(db.overrides).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get active overrides")
	}
	defer rows.Close()

	var overrides []override.UserOverride
	for rows.Next() {
		var o override.UserOverride
		var schedTgt sql.NullString
		var add, rem sql.NullString
		err = rows.Scan(&add, &rem, &schedTgt)
		if err != nil {
			return errors.Wrap(err, "scan override")
		}
		o.AddUserID = add.String
		o.RemoveUserID = rem.String
		if !schedTgt.Valid {
			continue
		}
		o.Target = assignment.ScheduleTarget(schedTgt.String)
		overrides = append(overrides, o)
	}

	rows, err = tx.Stmt(db.rules).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get rules")
	}
	defer rows.Close()

	type userRule struct {
		rule.Rule
		UserID string
	}

	var rules []userRule
	var tzName string
	tz := make(map[string]*time.Location)
	for rows.Next() {
		var r userRule
		filter := make(pq.BoolArray, 7)
		err = rows.Scan(
			&r.ScheduleID,
			&filter,
			&r.Start,
			&r.End,
			&tzName,
			&r.UserID,
		)
		if err != nil {
			return errors.Wrap(err, "scan rule")
		}
		for i, v := range filter {
			r.SetDay(time.Weekday(i), v)
		}
		if tz[r.ScheduleID] == nil {
			tz[r.ScheduleID], err = util.LoadLocation(tzName)
			if err != nil {
				return errors.Wrap(err, "load timezone")
			}
		}

		rules = append(rules, r)
	}
	rows, err = tx.Stmt(db.getOnCall).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get on call")
	}
	defer rows.Close()

	type onCall struct {
		UserID     string
		ScheduleID string
	}
	oldOnCall := make(map[onCall]bool)
	var oc onCall
	for rows.Next() {
		err = rows.Scan(&oc.ScheduleID, &oc.UserID)
		if err != nil {
			return errors.Wrap(err, "scan on call user")
		}
		oldOnCall[oc] = true
	}

	newOnCall := make(map[onCall]bool, len(rules))
	for _, r := range rules {
		if r.IsActive(now.In(tz[r.ScheduleID])) {
			newOnCall[onCall{ScheduleID: r.ScheduleID, UserID: r.UserID}] = true
		}
	}

	for _, o := range overrides {
		if o.AddUserID != "" && o.RemoveUserID == "" {
			// ADD override
			newOnCall[onCall{ScheduleID: o.Target.TargetID(), UserID: o.AddUserID}] = true
			continue
		}
		if o.AddUserID == "" && o.RemoveUserID != "" {
			// REMOVE override
			delete(newOnCall, onCall{ScheduleID: o.Target.TargetID(), UserID: o.RemoveUserID})
			continue
		}

		if newOnCall[onCall{ScheduleID: o.Target.TargetID(), UserID: o.RemoveUserID}] {
			// REPLACE override
			delete(newOnCall, onCall{ScheduleID: o.Target.TargetID(), UserID: o.RemoveUserID})
			newOnCall[onCall{ScheduleID: o.Target.TargetID(), UserID: o.AddUserID}] = true
		}
	}

	start := tx.Stmt(db.startOnCall)

	for oc := range newOnCall {
		// not on call in DB, but are now
		if !oldOnCall[oc] {
			_, err = start.ExecContext(ctx, oc.ScheduleID, oc.UserID)
			if err != nil && !isScheduleDeleted(err) {
				return errors.Wrap(err, "record shift start")
			}
		}
	}
	end := tx.Stmt(db.endOnCall)
	for oc := range oldOnCall {
		// on call in DB, but no longer
		if !newOnCall[oc] {
			_, err = end.ExecContext(ctx, oc.ScheduleID, oc.UserID)
			if err != nil {
				return errors.Wrap(err, "record shift end")
			}
		}
	}

	return tx.Commit()
}

func isScheduleDeleted(err error) bool {
	dbErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return dbErr.Constraint == "schedule_on_call_users_schedule_id_fkey"
}
