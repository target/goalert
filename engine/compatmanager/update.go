package compatmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/slack"
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

	q := gadb.New(tx)
	rows, err := q.CompatCMMissingSub(ctx, slack.DestTypeSlackDirectMessage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	for _, row := range rows {
		u, err := db.cs.User(ctx, row.Dest.DestV1.Arg(slack.FieldSlackUserID))
		if err != nil {
			log.Log(ctx, fmt.Errorf("update auth subjects: lookup Slack user (%s): %w", row.Dest.DestV1.Arg(slack.FieldSlackUserID), err))
			continue
		}

		err = q.CompatUpsertAuthSubject(ctx, gadb.CompatUpsertAuthSubjectParams{
			UserID:     row.UserID,
			ProviderID: "slack:" + u.TeamID,
			SubjectID:  u.ID,
			CmID:       uuid.NullUUID{UUID: row.ID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("upsert auth subject: %w", err)
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

	q := gadb.New(tx)
	rows, err := q.CompatSlackSubMissingCM(ctx)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	for _, s := range rows {
		// provider id contains the team id in the format "slack:team_id"
		// but we need to store the contact method id in the format "team_id:subject_id"
		teamID := strings.TrimPrefix(s.ProviderID, "slack:")
		value := s.SubjectID
		team, err := db.cs.Team(ctx, teamID)
		if err != nil {
			log.Log(ctx, err)
			continue
		}

		id := uuid.New()
		err = q.CompatInsertUserCM(ctx, gadb.CompatInsertUserCMParams{
			ID:   id,
			Name: team.Name,
			Dest: gadb.NullDestV1{
				DestV1: gadb.DestV1{
					Type: slack.DestTypeSlackDirectMessage,
					Args: map[string]string{
						slack.FieldSlackUserID: value,
					},
				},
				Valid: true,
			},
			UserID: s.UserID,
		})
		if err != nil {
			return fmt.Errorf("insert cm: %w", err)
		}

		err = q.CompatLinkAuthSubjectCM(ctx, gadb.CompatLinkAuthSubjectCMParams{
			ID:   s.ID,
			CmID: uuid.NullUUID{UUID: id, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("update sub cm_id: %w", err)
		}
	}

	return tx.Commit()
}
