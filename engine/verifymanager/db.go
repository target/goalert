package verifymanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB will manage verification codes.
type DB struct {
	lock *processinglock.Lock

	insertMessages *sql.Stmt
	cleanupExpired *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.VerificationManager" }

// NewDB will create a new DB instance, preparing all statements.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeVerify,
		Version: 2,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		lock: lock,
		insertMessages: p.P(`
			with rows as (
				insert into outgoing_messages (message_type, contact_method_id, user_id, user_verification_code_id)
				select 'verification_message', contact_method_id, cm.user_id, code.id
				from user_verification_codes code
				join user_contact_methods cm on cm.id = contact_method_id
				where not sent and now() < expires_at
				limit 100
				for update skip locked
				returning user_verification_code_id id
			)
			update user_verification_codes code
			set sent = true
			from rows
			where code.id = rows.id
		`),

		cleanupExpired: p.P(`
			with rows as (
				select id, contact_method_id
				from user_verification_codes
				where now() >= expires_at
				limit 100
				for update skip locked
			), _cms as (
				delete from user_contact_methods cm
				using rows
				where cm.id = rows.contact_method_id and cm.pending
			)
			delete from user_verification_codes code
			using rows
			where code.id = rows.id
		`),
	}, p.Err
}
