package message

import (
	"context"
	"database/sql"
	"time"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// DB implements a priority message sender using Postgres.
type DB struct {
	lock *processinglock.Lock

	c Config

	stuckMessages *sql.Stmt

	setSending *sql.Stmt

	lockStmt     *sql.Stmt
	messageQueue *sql.Stmt // todo: something that doesn't remind Aru of websphere
	currentTime  *sql.Stmt
	retryReset   *sql.Stmt
	retryClear   *sql.Stmt

	sendDeadlineExpired *sql.Stmt

	failDisabledCM *sql.Stmt

	sentByCMType *sql.Stmt

	updateCMStatusUpdate      *sql.Stmt
	cleanupStatusUpdateOptOut *sql.Stmt

	tempFail     *sql.Stmt
	permFail     *sql.Stmt
	updateStatus *sql.Stmt

	advLock        *sql.Stmt
	advLockCleanup *sql.Stmt
}

// NewDB creates a new DB. If config is nil, DefaultConfig() is used.
func NewDB(ctx context.Context, db *sql.DB, c *Config) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeMessage,
		Version: 6,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

	if c == nil {
		c = DefaultConfig()
	}
	err = validate.Range("MaxMessagesPerCycle", c.MaxMessagesPerCycle, 0, 9000)
	if err != nil {
		return nil, err
	}
	if c.MaxMessagesPerCycle == 0 {
		c.MaxMessagesPerCycle = 50
	}

	tempFail := p.P(`
		update outgoing_messages
		set
			last_status = 'failed',
			last_status_at = now(),
			status_details = $3,
			provider_msg_id = coalesce($2, provider_msg_id),
			next_retry_at = CASE WHEN retry_count < 3 THEN now() + '15 seconds'::interval ELSE null END
		where id = $1 or provider_msg_id = $2
	`)
	permFail := p.P(`
		update outgoing_messages
		set
			last_status = 'failed',
			last_status_at = now(),
			status_details = $3,
			cycle_id = null,
			provider_msg_id = coalesce($2, provider_msg_id),
			next_retry_at = null
		where id = $1 or provider_msg_id = $2
	`)
	updateStatus := p.P(`
		update outgoing_messages
		set
			last_status = cast($4 as enum_outgoing_messages_status),
			last_status_at = now(),
			status_details = $5,
			cycle_id = null,
			sending_deadline = null,
			sent_at = coalesce(sent_at, fired_at, now()),
			fired_at = null,
			provider_msg_id = coalesce($2, provider_msg_id),
			provider_seq = CASE WHEN $3 = -1 THEN provider_seq ELSE $3 END,
			next_retry_at = null
		where
			(id = $1 or provider_msg_id = $2) and
			(provider_seq <= $3 or $3 = -1) and
			last_status not in ('failed', 'pending')
	`)
	if p.Err != nil {
		return nil, p.Err
	}
	return &DB{
		lock: lock,
		c:    *c,

		updateStatus: updateStatus,
		tempFail:     tempFail,
		permFail:     permFail,

		advLock: p.P(`select pg_advisory_lock($1)`),
		advLockCleanup: p.P(`
			select pg_terminate_backend(lock.pid)
			from pg_locks lock
			join pg_database pgdat on
				datname = current_database() and
				lock.database = pgdat.oid
			join pg_stat_activity act on
				act.datid = pgdat.oid and
				act.pid = lock.pid and
				act.state = 'idle' and
				act.state_change < now() - '1 minute'::interval
			where objid = $1 and locktype = 'advisory' and granted
		`),

		stuckMessages: p.P(`
			with sel as (
				select id, provider_msg_id
				from outgoing_messages msg
				where
					last_status = 'queued_remotely' and
					last_status_at < now()-'1 minute'::interval and
					provider_msg_id notnull
				order by
					last_status_at
				limit 10
				for update
			)
			update outgoing_messages msg
			set last_status_at = now()
			from sel
			where msg.id = sel.id
			returning msg.id, msg.provider_msg_id
		`),

		sentByCMType: p.P(`
			select count(*)
			from outgoing_messages msg
			join user_contact_methods cm on cm.id = msg.contact_method_id
			where msg.sent_at > $1 and cm.type = $2
		`),

		updateCMStatusUpdate: p.P(`
			update outgoing_messages msg
			set contact_method_id = usr.alert_status_log_contact_method_id
			from users usr
			where
				msg.message_type = 'alert_status_update' and
				(
					msg.last_status = 'pending' or
					(msg.last_status = 'failed' and msg.next_retry_at notnull)
				) and
				msg.contact_method_id != usr.alert_status_log_contact_method_id and
				msg.user_id = usr.id and
				usr.alert_status_log_contact_method_id notnull
		`),
		cleanupStatusUpdateOptOut: p.P(`
			delete from outgoing_messages msg
			using users usr
			where
				msg.message_type = 'alert_status_update' and
				(
					msg.last_status = 'pending' or
					(msg.last_status = 'failed' and msg.next_retry_at notnull)
				) and
				usr.alert_status_log_contact_method_id isnull and
				usr.id = msg.user_id
		`),
		setSending: p.P(`
			update outgoing_messages
			set
				last_status = 'sending',
				last_status_at = now(),
				status_details = '',
				sending_deadline = now() + '10 seconds'::interval,
				fired_at = now(),
				provider_seq = 0,
				provider_msg_id = null,
				next_retry_at = null
			where id = $1
		`),

		sendDeadlineExpired: p.P(`
			update outgoing_messages
			set
				last_status = 'failed',
				last_status_at = now(),
				status_details = 'send deadline expired',
				cycle_id = null,
				next_retry_at = null
			where
				last_status = 'sending' and
				sending_deadline <= now()
		`),
		retryReset: p.P(`
			update outgoing_messages
			set
				last_status = 'pending',
				status_details = '',
				next_retry_at = null,
				retry_count = retry_count + 1,
				fired_at = null,
				sent_at = null,
				provider_msg_id = null,
				provider_seq = 0
			where
				last_status = 'failed' and
				now() > next_retry_at and
				retry_count < 3
		`),
		retryClear: p.P(`
			update outgoing_messages
			set
				next_retry_at = null,
				cycle_id = null
			where
				last_status = 'failed' and
				retry_count >= 3 and
				(cycle_id notnull or next_retry_at notnull)
		`),

		lockStmt:    p.P(`lock outgoing_messages in exclusive mode`),
		currentTime: p.P(`select now()`),

		failDisabledCM: p.P(`
			update outgoing_messages msg
			set
				last_status = 'failed',
				last_status_at = now(),
				status_details = 'contact method disabled',
				cycle_id = null,
				next_retry_at = null
			from user_contact_methods cm
			where
				msg.last_status = 'pending' and
				msg.message_type != 'verification_message' and
				cm.id = msg.contact_method_id and
				cm.disabled
		`),

		messageQueue: p.P(`
			select
				msg.id,
				msg.message_type,
				cm.type,
				chan.type,
				coalesce(msg.contact_method_id, msg.channel_id),
				coalesce(cm.value, chan.value),
				msg.alert_id,
				msg.alert_log_id,
				msg.user_verification_code_id,
				msg.user_id,
				msg.service_id,
				msg.created_at,
				msg.sent_at
			from outgoing_messages msg
			left join user_contact_methods cm on cm.id = msg.contact_method_id
			left join notification_channels chan on chan.id = msg.channel_id
			where
				sent_at > now() - '10 minutes'::interval or
				last_status = 'pending' and
				(msg.contact_method_id isnull or msg.message_type = 'verification_message' or not cm.disabled)
		`), // TODO: base time after historical data

	}, p.Err
}

func (db *DB) currentQueue(ctx context.Context, tx *sql.Tx, now time.Time) (*queue, error) {
	rows, err := tx.Stmt(db.messageQueue).QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch outgoing messages")
	}
	defer rows.Close()

	var result []Message
	for rows.Next() {
		var msg Message
		var cID, cValue, verifyID, userID, serviceID, cmType, chanType sql.NullString
		var alertID, logID sql.NullInt64
		var createdAt, sentAt sql.NullTime
		err = rows.Scan(
			&msg.ID,
			&msg.Type,
			&cmType,
			&chanType,
			&cID,
			&cValue,
			&alertID,
			&logID,
			&verifyID,
			&userID,
			&serviceID,
			&createdAt,
			&sentAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan row")
		}
		msg.AlertID = int(alertID.Int64)
		msg.AlertLogID = int(logID.Int64)
		msg.VerifyID = verifyID.String
		msg.UserID = userID.String
		msg.ServiceID = serviceID.String
		msg.CreatedAt = createdAt.Time
		msg.SentAt = sentAt.Time
		msg.Dest.ID = cID.String
		msg.Dest.Value = cValue.String
		switch {
		case cmType.String == string(contactmethod.TypeSMS):
			msg.Dest.Type = notification.DestTypeSMS
		case cmType.String == string(contactmethod.TypeVoice):
			msg.Dest.Type = notification.DestTypeVoice
		case chanType.String == string(notificationchannel.TypeSlack):
			msg.Dest.Type = notification.DestTypeSlackChannel
		default:
			log.Debugf(ctx, "unknown message type for message %s", msg.ID)
			continue
		}

		result = append(result, msg)
	}

	return newQueue(result, now), nil
}

// UpdateMessageStatus will update the state of a message.
func (db *DB) UpdateMessageStatus(ctx context.Context, status *notification.MessageStatus) error {
	return retry.DoTemporaryError(func(int) error {
		return db._UpdateMessageStatus(ctx, status)
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Millisecond*100),
	)
}
func (db *DB) _UpdateMessageStatus(ctx context.Context, status *notification.MessageStatus) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	var cbID, pID sql.NullString
	if status.ID != "" {
		cbID.Valid = true
		cbID.String = status.ID
	}
	if status.ProviderMessageID != "" {
		pID.Valid = true
		pID.String = status.ProviderMessageID
	}

	if status.State == notification.MessageStateFailedTemp {
		_, err = db.tempFail.ExecContext(ctx, cbID, pID, status.Details)
		return err
	}
	if status.State == notification.MessageStateFailedPerm {
		_, err = db.permFail.ExecContext(ctx, cbID, pID, status.Details)
		return err
	}

	var s Status
	switch status.State {
	case notification.MessageStateActive:
		s = StatusQueuedRemotely
	case notification.MessageStateSent:
		s = StatusSent
	case notification.MessageStateDelivered:
		s = StatusDelivered
	}

	_, err = db.updateStatus.ExecContext(ctx, cbID, pID, status.Sequence, s, status.Details)
	return err
}
