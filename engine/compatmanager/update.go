package compatmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// UpdateAll will process compatibility entries for the cycle.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running compat operations.")

	err = db.updateContactMethods(ctx)
	if err != nil {
		return fmt.Errorf("update contact methods: %w", err)
	}

	err = db.updateAuthSubjects(ctx)
	if err != nil {
		return fmt.Errorf("update auth subjects: %w", err)
	}

	return nil
}

func (db *DB) updateAuthSubjects(ctx context.Context) error {
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "engine: update auth subjects", tx)

	type cm struct {
		ID          uuid.UUID
		UserID      uuid.UUID
		SlackUserID string
		SlackTeamID string
	}

	var cms []cm
	rows, err := tx.StmtContext(ctx, db.cmMissingSub).QueryContext(ctx)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	for rows.Next() {
		var c cm
		err = rows.Scan(&c.ID, &c.UserID, &c.SlackUserID)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		u, err := db.cs.User(ctx, c.SlackUserID)
		if err != nil {
			log.Log(ctx, fmt.Errorf("update auth subjects: lookup Slack user (%s): %w", c.SlackUserID, err))
			continue
		}

		c.SlackTeamID = u.TeamID
		cms = append(cms, c)
	}

	for _, c := range cms {
		_, err = tx.StmtContext(ctx, db.insertSub).ExecContext(ctx, c.UserID, c.SlackUserID, "slack:"+c.SlackTeamID, c.ID)
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func (db *DB) updateContactMethods(ctx context.Context) error {
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "engine: update contact methods", tx)

	type sub struct {
		ID         int
		UserID     string
		SubjectID  string
		ProviderID string
	}

	var subs []sub
	rows, err := tx.StmtContext(ctx, db.slackSubMissingCM).QueryContext(ctx)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	for rows.Next() {
		var s sub
		err = rows.Scan(&s.ID, &s.UserID, &s.SubjectID, &s.ProviderID)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		subs = append(subs, s)
	}

	for _, s := range subs {
		// provider id contains the team id in the format "slack:team_id"
		// but we need to store the contact method id in the format "team_id:subject_id"
		teamID := strings.TrimPrefix(s.ProviderID, "slack:")
		value := s.SubjectID
		name, err := db.cs.TeamName(ctx, teamID)
		if err != nil {
			log.Log(ctx, err)
			continue
		}

		_, err = tx.StmtContext(ctx, db.insertCM).ExecContext(ctx, uuid.New(), name, "SLACK_DM", value, s.UserID)
		if err != nil {
			return fmt.Errorf("insert cm: %w", err)
		}

		_, err = tx.StmtContext(ctx, db.updateSubCMID).ExecContext(ctx, s.ID, value)
		if err != nil {
			return fmt.Errorf("update sub cm_id: %w", err)
		}
	}

	return tx.Commit()
}
