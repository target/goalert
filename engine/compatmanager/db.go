package compatmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/util"
)

// DB handles keeping compatibility-related data in sync.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	cs *slack.ChannelSender

	slackSubMissingCM *sql.Stmt
	updateSubCMID     *sql.Stmt
	insertCM          *sql.Stmt

	cmMissingSub *sql.Stmt
	insertSub    *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CompatManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, cs *slack.ChannelSender) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeCompat,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,
		cs:   cs,

		// get all entries missing cm_id where provider_id starts with "slack:"
		slackSubMissingCM: p.P(`
			select id, user_id, subject_id, provider_id from auth_subjects where
				provider_id like 'slack:%' and cm_id is null
			for update skip locked
			limit 10
		`),

		// update cm_id for a given user_id and subject_id
		updateSubCMID: p.P(`
			update auth_subjects
			set cm_id = (
				select id from user_contact_methods
				where type = 'SLACK_DM' and value = $2
			) where id = $1
		`),

		insertCM: p.P(`
			insert into user_contact_methods (id, name, type, value, user_id, pending)
			values ($1, $2, $3, $4, $5, false)
			on conflict (type, value) do nothing
		`),

		// find verified contact methods (disabled false) with no auth subject
		cmMissingSub: p.P(`
			select id, user_id, value from user_contact_methods where
			type = 'SLACK_DM' and not disabled and not exists (
				select 1 from auth_subjects where cm_id = user_contact_methods.id
			)
			for update skip locked
			limit 10
		`),

		insertSub: p.P(`
			insert into auth_subjects (user_id, subject_id, provider_id, cm_id)
			values ($1, $2, $3, $4)
			on conflict (subject_id, provider_id) do update set user_id = $1, cm_id = $4
		`),
	}, p.Err
}
