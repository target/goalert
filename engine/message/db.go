package message

import (
	"context"
	"database/sql"
	"sync"
	"time"

	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
	"go.opencensus.io/trace"

	"github.com/pkg/errors"
)

// DB implements a priority message sender using Postgres.
type DB struct {
	lock *processinglock.Lock

	c Config

	stuckMessages *sql.Stmt

	setSending *sql.Stmt

	lockStmt    *sql.Stmt
	messages    *sql.Stmt
	currentTime *sql.Stmt
	retryReset  *sql.Stmt
	retryClear  *sql.Stmt

	sendDeadlineExpired *sql.Stmt

	failDisabledCM *sql.Stmt
	alertlogstore  alertlog.Store

	failSMSVoice *sql.Stmt

	sentByCMType *sql.Stmt

	updateCMStatusUpdate      *sql.Stmt
	cleanupStatusUpdateOptOut *sql.Stmt

	tempFail     *sql.Stmt
	permFail     *sql.Stmt
	updateStatus *sql.Stmt

	advLock        *sql.Stmt
	advLockCleanup *sql.Stmt

	insertAlertBundle  *sql.Stmt
	insertStatusBundle *sql.Stmt
}

// NewDB creates a new DB. If config is nil, DefaultConfig() is used.
func NewDB(ctx context.Context, db *sql.DB, c *Config, a alertlog.Store) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeMessage,
		Version: 7,
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
		lock:          lock,
		c:             *c,
		alertlogstore: a,

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
			with disabled as (
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
				returning msg.id as msg_id, alert_id, msg.user_id, cm.id as cm_id
			) select distinct msg_id, alert_id, user_id, cm_id from disabled where alert_id notnull
		`),

		failSMSVoice: p.P(`
			update outgoing_messages msg
			set
				last_status = 'failed',
				last_status_at = now(),
				status_details = 'SMS/Voice support not enabled by administrator',
				cycle_id = null,
				next_retry_at = null
			from user_contact_methods cm
			where
				msg.last_status = 'pending' and
				cm.type in ('SMS', 'VOICE') and
				cm.id = msg.contact_method_id
			returning msg.id as msg_id, alert_id, msg.user_id, cm.id as cm_id
		`),

		insertAlertBundle: p.P(`
			with new_msg as (
				insert into outgoing_messages (
					id,
					created_at,
					message_type,
					contact_method_id,
					channel_id,
					user_id,
					service_id
				) values (
					$1, $2, 'alert_notification_bundle', $3, $4, $5, $6
				) returning (id)
			)
			update outgoing_messages
			set
				last_status = 'bundled',
				last_status_at = now(),
				status_details = (select id from new_msg),
				cycle_id = null
			where id = any($7::uuid[])
		`),

		insertStatusBundle: p.P(`
			with new_msg as (
				insert into outgoing_messages (
					id,
					created_at,
					message_type,
					contact_method_id,
					user_id,
					alert_log_id,
					status_alert_ids
				) values (
					$1, $2, 'alert_status_update_bundle', $3, $4, $5, $6::bigint[]
				) returning (id)
			)
			update outgoing_messages
			set
				last_status = 'bundled',
				last_status_at = now(),
				status_details = (select id from new_msg)
			where id = any($7::uuid[])
		`),

		messages: p.P(`
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
				cm.user_id,
				msg.service_id,
				msg.created_at,
				msg.sent_at,
				msg.status_alert_ids
			from outgoing_messages msg
			left join user_contact_methods cm on cm.id = msg.contact_method_id
			left join notification_channels chan on chan.id = msg.channel_id
			where
				sent_at > now() - '10 minutes'::interval or
				last_status = 'pending' and
				(msg.contact_method_id isnull or msg.message_type = 'verification_message' or not cm.disabled)
		`),
	}, p.Err
}

func (db *DB) currentQueue(ctx context.Context, tx *sql.Tx, now time.Time) (*queue, error) {
	rows, err := tx.Stmt(db.messages).QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch outgoing messages")
	}
	defer rows.Close()

	var result []Message
	for rows.Next() {
		var msg Message
		var destID, destValue, verifyID, userID, serviceID, cmType, chanType sql.NullString
		var alertID, logID sql.NullInt64
		var statusAlertIDs sqlutil.IntArray
		var createdAt, sentAt sql.NullTime
		err = rows.Scan(
			&msg.ID,
			&msg.Type,
			&cmType,
			&chanType,
			&destID,
			&destValue,
			&alertID,
			&logID,
			&verifyID,
			&userID,
			&serviceID,
			&createdAt,
			&sentAt,
			&statusAlertIDs,
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
		msg.Dest.ID = destID.String
		msg.Dest.Value = destValue.String
		msg.StatusAlertIDs = statusAlertIDs
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

	cfg := config.FromContext(ctx)
	if cfg.General.MessageBundles {
		result, err = bundleStatusMessages(result, func(msg Message, ids []string) error {
			_, err := tx.StmtContext(ctx, db.insertStatusBundle).ExecContext(ctx, msg.ID, msg.CreatedAt, msg.Dest.ID, msg.UserID, msg.AlertLogID, sqlutil.IntArray(msg.StatusAlertIDs), sqlutil.UUIDArray(ids))
			return errors.Wrap(err, "insert status bundle")
		})
		if err != nil {
			return nil, err
		}
		result, err = bundleAlertMessages(result, func(msg Message, ids []string) error {
			var cmID, chanID, userID sql.NullString
			if msg.UserID != "" {
				userID.Valid = true
				userID.String = msg.UserID
			}
			if msg.Dest.Type.IsUserCM() {
				cmID.Valid = true
				cmID.String = msg.Dest.ID
			} else {
				chanID.Valid = true
				chanID.String = msg.Dest.ID
			}
			_, err := tx.StmtContext(ctx, db.insertAlertBundle).ExecContext(ctx, msg.ID, msg.CreatedAt, cmID, chanID, userID, msg.ServiceID, sqlutil.UUIDArray(ids))
			return err
		})
		if err != nil {
			return nil, err
		}
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
	case notification.MessageStateSending:
		s = StatusQueuedRemotely
	case notification.MessageStateSent:
		s = StatusSent
	case notification.MessageStateDelivered:
		s = StatusDelivered
	}

	_, err = db.updateStatus.ExecContext(ctx, cbID, pID, status.Sequence, s, status.Details)
	return err
}

// SendFunc defines a function that sends messages.
type SendFunc func(context.Context, *Message) (*notification.MessageStatus, error)

// ErrAbort is returned when an early-abort is returned due to pause.
var ErrAbort = errors.New("aborted due to pause")

// StatusFunc is used to fetch the latest status of a message.
type StatusFunc func(ctx context.Context, id, providerMsgID string) (*notification.MessageStatus, error)

// SendMessages will send notifications using SendFunc.
func (db *DB) SendMessages(ctx context.Context, send SendFunc, status StatusFunc) error {
	err := db._SendMessages(ctx, send, status)
	if db.c.Pausable.IsPausing() {
		return ErrAbort
	}
	return err
}

func (db *DB) _SendMessages(ctx context.Context, send SendFunc, status StatusFunc) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Sending outgoing messages.")

	execCtx, execCancel := context.WithCancel(ctx)
	execDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-db.c.Pausable.PauseWait():
		case <-execDone:
		}
		execCancel()
	}()

	res, err := db.advLockCleanup.ExecContext(execCtx, lock.GlobalMessageSending)
	if err != nil {
		return errors.Wrap(err, "terminate stale backend locks")
	}
	rowsCount, _ := res.RowsAffected()
	if rowsCount > 0 {
		log.Log(execCtx, errors.Errorf("terminated %d stale backend instance(s) holding message sending lock", rowsCount))
	}

	cLock, err := db.lock.Conn(execCtx)
	if err != nil {
		return errors.Wrap(err, "get DB conn")
	}
	defer cLock.Close()

	_, err = cLock.Exec(execCtx, db.advLock, lock.GlobalMessageSending)
	if err != nil {
		return errors.Wrap(err, "acquire global sending advisory lock")
	}
	defer func() {
		ctx := trace.NewContext(context.Background(), trace.FromContext(execCtx))
		cLock.ExecWithoutLock(ctx, `select pg_advisory_unlock_all()`)
	}()

	tx, err := cLock.BeginTx(execCtx, nil)
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}
	defer tx.Rollback()

	_, err = tx.Stmt(db.lockStmt).ExecContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "acquire exclusive locks")
	}

	var t time.Time
	err = tx.Stmt(db.currentTime).QueryRowContext(execCtx).Scan(&t)
	if err != nil {
		return errors.Wrap(err, "get current time")
	}

	_, err = tx.Stmt(db.updateCMStatusUpdate).ExecContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "update status update CM preferences")
	}

	_, err = tx.Stmt(db.cleanupStatusUpdateOptOut).ExecContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "clear disabled status updates")
	}

	type msgMeta struct {
		MessageID string
		AlertID   int
		UserID    string
		CMID      string
	}

	var msgs []msgMeta

	// if twilio is disable, create an entry to notify the user
	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		rows, err := tx.StmtContext(ctx, db.failSMSVoice).QueryContext(execCtx)
		if err != nil {
			return errors.Wrap(err, "check for failed message")
		}
		defer rows.Close()

		for rows.Next() {
			var alertID sql.NullInt64
			var msg msgMeta
			err = rows.Scan(&msg.MessageID, &alertID, &msg.UserID, &msg.CMID)
			if err != nil {
				return errors.Wrap(err, "scan all failed messages")
			}
			if !alertID.Valid {
				continue
			}
			msg.AlertID = int(alertID.Int64)
			msgs = append(msgs, msg)
		}
	}

	// processes disabled CMs and writes to alert log if disabled
	rows, err := tx.Stmt(db.failDisabledCM).QueryContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "check for disabled CMs")
	}
	defer rows.Close()

	for rows.Next() {
		var msg msgMeta
		err = rows.Scan(&msg.MessageID, &msg.AlertID, &msg.UserID, &msg.CMID)
		if err != nil {
			return errors.Wrap(err, "scan all disabled CM messages")
		}
		msgs = append(msgs, msg)
	}

	for _, m := range msgs {
		meta := alertlog.NotificationMetaData{
			MessageID: m.MessageID,
		}

		// log failures
		db.alertlogstore.MustLogTx(permission.UserSourceContext(ctx, m.UserID, permission.RoleUser, &permission.SourceInfo{
			Type: permission.SourceTypeContactMethod,
			ID:   m.CMID,
		}), tx, m.AlertID, alertlog.TypeNotificationSent, meta)
	}

	_, err = tx.Stmt(db.sendDeadlineExpired).ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "fail expired messages")
	}

	_, err = tx.Stmt(db.retryClear).ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "clear max retries")
	}

	_, err = tx.Stmt(db.retryReset).ExecContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "reset retry messages")
	}

	q, err := db.currentQueue(ctx, tx, t)
	if err != nil {
		return errors.Wrap(err, "get pending messages")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit message updates")
	}

	var wg sync.WaitGroup
	for _, t := range q.Types() {
		wg.Add(1)
		go func(typ notification.DestType) {
			defer wg.Done()
			err := db.sendMessagesByType(ctx, cLock, send, q, typ)
			if err != nil && errors.Cause(err) != processinglock.ErrNoLock {
				log.Log(ctx, errors.Wrap(err, "send"))
			}
		}(t)
	}
	wg.Wait()

	return db.updateStuckMessages(ctx, status)
}

func (db *DB) refreshMessageState(ctx context.Context, statusFn StatusFunc, id, providerMsgID string, res chan *notification.MessageStatus) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status, err := statusFn(ctx, id, providerMsgID)
	if err != nil {
		res <- &notification.MessageStatus{
			Ctx:               ctx,
			ID:                id,
			ProviderMessageID: providerMsgID,
			State:             notification.MessageStateSending,
			Details:           "failed to update status: " + err.Error(),
			Sequence:          -1,
		}
		return
	}
	stat := *status
	if stat.State == notification.MessageStateFailedTemp {
		stat.State = notification.MessageStateFailedPerm
	}
	stat.Sequence = -1
	stat.Ctx = ctx
	res <- &stat
}
func (db *DB) updateStuckMessages(ctx context.Context, statusFn StatusFunc) error {
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Stmt(db.stuckMessages).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	type msg struct{ id, pID string }
	var toCheck []msg
	for rows.Next() {
		var m msg
		err = rows.Scan(&m.id, &m.pID)
		if err != nil {
			return err
		}
		toCheck = append(toCheck, m)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	ch := make(chan *notification.MessageStatus, len(toCheck))
	for _, m := range toCheck {
		go db.refreshMessageState(ctx, statusFn, m.id, m.pID, ch)
	}

	for range toCheck {
		err := db._UpdateMessageStatus(ctx, <-ch)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "update stale message status"))
		}
	}

	return nil
}

const workersPerType = 5

func (db *DB) sendMessagesByType(ctx context.Context, cLock *processinglock.Conn, send SendFunc, q *queue, typ notification.DestType) error {
	limit := db.c.RateLimit[typ]

	toSend := int(limit.Batch.Seconds() * float64(limit.PerSecond))
	var failures, processing int
	sentCount := func() int {
		return q.SentByType(typ, limit.Batch) - failures
	}

	sendCh := make(chan *Message, workersPerType)
	errCh := make(chan error, workersPerType)
	resCh := make(chan bool, workersPerType)
	for i := 0; i < workersPerType; i++ {
		go func() {
			for msg := range sendCh {
				sent, err := db.sendMessage(ctx, cLock, send, msg)
				if err != nil {
					errCh <- err
				} else {
					resCh <- sent
				}
			}
		}()
	}
	defer close(sendCh)

	var done bool
	for {
		if !done && sentCount() < toSend {
			msg := q.NextByType(typ)
			if msg == nil {
				done = true
				continue
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sendCh <- msg:
				processing++
			}
		} else if processing == 0 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case sent := <-resCh:
			if !sent {
				failures++
			}
			processing--
		}
	}

	return nil
}

func (db *DB) sendMessage(ctx context.Context, cLock *processinglock.Conn, send SendFunc, m *Message) (bool, error) {
	ctx, sp := trace.StartSpan(ctx, "Engine.MessageManager.SendMessage")
	defer sp.End()
	ctx = log.WithFields(ctx, log.Fields{
		"DestTypeID": m.Dest.ID,
		"DestType":   m.Dest.Type.String(),
		"CallbackID": m.ID,
	})
	sp.AddAttributes(
		trace.StringAttribute("message.dest.id", m.Dest.ID),
		trace.StringAttribute("message.dest.type", m.Dest.Type.String()),
		trace.StringAttribute("message.callback.id", m.ID),
	)
	if m.AlertID != 0 {
		ctx = log.WithField(ctx, "AlertID", m.AlertID)
	}
	_, err := cLock.Exec(ctx, db.setSending, m.ID)
	if err != nil {
		return false, err
	}
	sCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	var status *notification.MessageStatus
	err = retry.DoTemporaryError(func(int) error {
		status, err = send(sCtx, m)
		return err
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(65*time.Millisecond),
	)
	cancel()

	var pID sql.NullString
	if status != nil && status.ProviderMessageID != "" {
		pID.Valid = true
		pID.String = status.ProviderMessageID
	}

	retryExec := func(s *sql.Stmt, args ...interface{}) error {
		return retry.DoTemporaryError(func(int) error {
			_, err := s.ExecContext(ctx, args...)
			return err
		},
			retry.Limit(15),
			retry.FibBackoff(time.Millisecond*50),
		)
	}
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "send message"))

		err = retryExec(db.tempFail, m.ID, pID, err.Error())
		return false, errors.Wrap(err, "mark failed message")
	}

	if status.State == notification.MessageStateFailedTemp {
		err = retryExec(db.tempFail, m.ID, pID, status.Details)
		return false, errors.Wrap(err, "mark failed message (temp)")
	}
	if status.State == notification.MessageStateFailedPerm {
		err = retryExec(db.permFail, m.ID, pID, status.Details)
		return false, errors.Wrap(err, "mark failed message (perm)")
	}

	return true, errors.Wrap(db.UpdateMessageStatus(ctx, status), "update message status")
}
