package message

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"

	"github.com/pkg/errors"
)

// DB implements a priority message sender using Postgres.
type DB struct {
	lock *processinglock.Lock

	pausable lifecycle.Pausable

	stuckMessages *sql.Stmt

	setSending *sql.Stmt

	lockStmt    *sql.Stmt
	currentTime *sql.Stmt
	retryReset  *sql.Stmt
	retryClear  *sql.Stmt

	sendDeadlineExpired *sql.Stmt

	failDisabledCM *sql.Stmt
	alertlogstore  *alertlog.Store

	failSMSVoice *sql.Stmt

	sentByCMType *sql.Stmt

	cleanupStatusUpdateOptOut *sql.Stmt

	tempFail     *sql.Stmt
	permFail     *sql.Stmt
	updateStatus *sql.Stmt

	advLock        *sql.Stmt
	advLockCleanup *sql.Stmt

	createAlertBundle *sql.Stmt
	bundleMessages    *sql.Stmt

	deleteAny *sql.Stmt

	lastSent     time.Time
	sentMessages map[string]Message
}

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, a *alertlog.Store, pausable lifecycle.Pausable) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeMessage,
		Version: 10,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

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
			next_retry_at = null,
			src_value = coalesce(src_value, $6)
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
		pausable:      pausable,
		alertlogstore: a,

		updateStatus: updateStatus,
		tempFail:     tempFail,
		permFail:     permFail,

		sentMessages: make(map[string]Message),

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

		cleanupStatusUpdateOptOut: p.P(`
			delete from outgoing_messages msg
			using user_contact_methods cm
			where
				msg.message_type = 'alert_status_update' and
				(
					msg.last_status = 'pending' or
					(msg.last_status = 'failed' and msg.next_retry_at notnull)
				) and
				not cm.enable_status_updates and cm.id = msg.contact_method_id
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

		createAlertBundle: p.P(`
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
			)
		`),

		bundleMessages: p.P(`
			update outgoing_messages
			set
				last_status = 'bundled',
				last_status_at = now(),
				status_details = $1,
				cycle_id = null
			where id = any($2::uuid[])
		`),

		deleteAny: p.P(`delete from outgoing_messages where id = any($1)`),
	}, p.Err
}

func (db *DB) currentQueue(ctx context.Context, tx *sql.Tx, now time.Time) (*queue, error) {
	cutoff := now.Add(-maxThrottleDuration(PerCMThrottle, GlobalCMThrottle))
	sentSince := db.lastSent
	if sentSince.IsZero() {
		sentSince = cutoff
	}

	rows, err := gadb.New(tx).MessageMgrGetPending(ctx, sql.NullTime{Time: sentSince, Valid: true})
	if err != nil {
		return nil, errors.Wrap(err, "fetch outgoing messages")
	}

	result := make([]Message, 0, len(rows))
	for _, row := range rows {
		var msg Message
		msg.ID = row.ID.String()
		msg.Type = row.MessageType

		msg.AlertID = int(row.AlertID.Int64)
		msg.AlertLogID = int(row.AlertLogID.Int64)
		if row.UserVerificationCodeID.Valid {
			msg.VerifyID = row.UserVerificationCodeID.UUID.String()
		}
		if row.UserID.Valid {
			msg.UserID = row.UserID.UUID.String()
		}
		if row.ServiceID.Valid {
			msg.ServiceID = row.ServiceID.UUID.String()
		}
		msg.CreatedAt = row.CreatedAt
		msg.SentAt = row.SentAt.Time
		msg.Dest = notification.SQLDest{
			CMID:    row.CmID,
			CMType:  row.CmType,
			CMValue: row.CmValue,
			NCID:    row.ChanID,
			NCType:  row.ChanType,
			NCValue: row.ChanValue,
		}.Dest()
		msg.StatusAlertIDs = row.StatusAlertIds
		if row.ScheduleID.Valid {
			msg.ScheduleID = row.ScheduleID.UUID.String()
		}
		if msg.Dest.Type == notification.DestTypeUnknown {
			log.Debugf(ctx, "unknown message type for message %s", msg.ID)
			continue
		}

		if !msg.SentAt.IsZero() {
			// if the message was sent, just add it to the map
			db.sentMessages[msg.ID] = msg
			continue
		}

		result = append(result, msg)
	}

	for id, msg := range db.sentMessages {
		if msg.SentAt.Before(cutoff) {
			delete(db.sentMessages, id)
			continue
		}
		result = append(result, msg)
	}
	db.lastSent = now

	result, toDelete := dedupOnCallNotifications(result)
	if len(toDelete) > 0 {
		_, err = tx.StmtContext(ctx, db.deleteAny).ExecContext(ctx, sqlutil.UUIDArray(toDelete))
		if err != nil {
			return nil, fmt.Errorf("delete duplicate on-call notifications: %w", err)
		}
	}

	cfg := config.FromContext(ctx)
	result, toDelete = dedupStatusMessages(result)
	if len(toDelete) > 0 {
		_, err = tx.StmtContext(ctx, db.deleteAny).ExecContext(ctx, sqlutil.UUIDArray(toDelete))
		if err != nil {
			return nil, fmt.Errorf("delete duplicate status updates: %w", err)
		}
	}

	result, err = dedupAlerts(result, func(parentID string, duplicateIDs []string) error {
		_, err = tx.StmtContext(ctx, db.bundleMessages).ExecContext(ctx, parentID, sqlutil.UUIDArray(duplicateIDs))
		if err != nil {
			return fmt.Errorf("bundle '%v' by pointing to '%s': %w", duplicateIDs, parentID, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("dedup alerts: %w", err)
	}

	if cfg.General.DisableMessageBundles {
		return newQueue(result, now), nil
	}

	result, err = bundleAlertMessages(result, func(msg Message) (string, error) {
		var cmID, chanID uuid.NullUUID
		var userID sql.NullString
		if msg.UserID != "" {
			userID.Valid = true
			userID.String = msg.UserID
		}
		if msg.Dest.ID.IsUserCM() {
			cmID.Valid = true
			cmID.UUID = msg.Dest.ID.UUID()
		} else {
			chanID.Valid = true
			chanID.UUID = msg.Dest.ID.UUID()
		}

		newID := uuid.NewString()
		_, err := tx.StmtContext(ctx, db.createAlertBundle).ExecContext(ctx, newID, msg.CreatedAt, cmID, chanID, userID, msg.ServiceID)
		if err != nil {
			return "", err
		}

		return newID, nil
	}, func(parentID string, ids []string) error {
		_, err = tx.StmtContext(ctx, db.bundleMessages).ExecContext(ctx, parentID, sqlutil.UUIDArray(ids))
		return err
	})
	if err != nil {
		return nil, err
	}

	return newQueue(result, now), nil
}

// UpdateMessageStatus will update the state of a message.
func (db *DB) UpdateMessageStatus(ctx context.Context, status *notification.SendResult) error {
	return retry.DoTemporaryError(func(int) error {
		return db._UpdateMessageStatus(ctx, status)
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Millisecond*100),
	)
}

func (db *DB) _UpdateMessageStatus(ctx context.Context, status *notification.SendResult) error {
	if status == nil {
		// nothing to do
		return nil
	}
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	var cbID sql.NullString
	if status.ID != "" {
		cbID.Valid = true
		cbID.String = status.ID
	}

	if status.State == notification.StateFailedTemp {
		_, err = db.tempFail.ExecContext(ctx, cbID, status.ProviderMessageID, status.Details)
		return err
	}
	if status.State == notification.StateFailedPerm {
		_, err = db.permFail.ExecContext(ctx, cbID, status.ProviderMessageID, status.Details)
		return err
	}

	var s Status
	switch status.State {
	case notification.StateSending:
		s = StatusQueuedRemotely
	case notification.StateSent:
		s = StatusSent
	case notification.StateDelivered:
		s = StatusDelivered
	}

	var srcValue sql.NullString
	if status.SrcValue != "" {
		srcValue.Valid = true
		srcValue.String = status.SrcValue
	}

	_, err = db.updateStatus.ExecContext(ctx, cbID, status.ProviderMessageID, status.Sequence, s, status.Details, srcValue)
	return err
}

// SendFunc defines a function that sends messages.
type SendFunc func(context.Context, *Message) (*notification.SendResult, error)

// ErrAbort is returned when an early-abort is returned due to pause.
var ErrAbort = errors.New("aborted due to pause")

// StatusFunc is used to fetch the latest status of a message.
type StatusFunc func(ctx context.Context, providerID notification.ProviderMessageID) (*notification.Status, notification.DestType, error)

// SendMessages will send notifications using SendFunc.
func (db *DB) SendMessages(ctx context.Context, send SendFunc, status StatusFunc) error {
	err := db._SendMessages(ctx, send, status)
	if db.pausable.IsPausing() {
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
		case <-db.pausable.PauseWait():
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
		_, _ = cLock.ExecWithoutLock(log.FromContext(execCtx).BackgroundContext(), `select pg_advisory_unlock(4912)`)
	}()

	tx, err := cLock.BeginTx(execCtx, nil)
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}
	defer sqlutil.Rollback(ctx, "engine: message: send", tx)

	_, err = tx.Stmt(db.lockStmt).ExecContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "acquire exclusive locks")
	}

	var t time.Time
	err = tx.Stmt(db.currentTime).QueryRowContext(execCtx).Scan(&t)
	if err != nil {
		return errors.Wrap(err, "get current time")
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
		go func(typ string) {
			defer wg.Done()
			err := db.sendMessagesByType(ctx, cLock, send, q, typ)
			if err != nil && !errors.Is(err, processinglock.ErrNoLock) {
				log.Log(ctx, errors.Wrap(err, "send"))
			}
		}(t)
	}
	wg.Wait()

	return db.updateStuckMessages(ctx, status)
}

func (db *DB) refreshMessageState(ctx context.Context, statusFn StatusFunc, id string, providerID notification.ProviderMessageID, res chan *notification.SendResult) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status, _, err := statusFn(ctx, providerID)
	if errors.Is(err, notification.ErrStatusUnsupported) {
		// not available
		res <- nil
		return
	}
	if err != nil {
		// failed, log error
		log.Log(ctx, err)
		res <- nil
		return
	}

	stat := *status
	if stat.State == notification.StateFailedTemp {
		stat.State = notification.StateFailedPerm
	}
	stat.Sequence = -1
	res <- &notification.SendResult{Status: stat, ID: id, ProviderMessageID: providerID}
}

func (db *DB) updateStuckMessages(ctx context.Context, statusFn StatusFunc) error {
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "message: update stuck messages", tx)

	rows, err := tx.Stmt(db.stuckMessages).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	type msg struct {
		id         string
		providerID notification.ProviderMessageID
	}
	var toCheck []msg
	for rows.Next() {
		var m msg
		err = rows.Scan(&m.id, &m.providerID)
		if err != nil {
			return err
		}
		toCheck = append(toCheck, m)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	ch := make(chan *notification.SendResult, len(toCheck))
	for _, m := range toCheck {
		go db.refreshMessageState(ctx, statusFn, m.id, m.providerID, ch)
	}

	for range toCheck {
		err := db._UpdateMessageStatus(ctx, <-ch)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "update stale message status"))
		}
	}

	return nil
}

func (db *DB) sendMessagesByType(ctx context.Context, cLock *processinglock.Conn, send SendFunc, q *queue, typ string) error {
	ch := make(chan error)
	var count int
	for {
		msg := q.NextByType(typ)
		if msg == nil {
			break
		}
		count++
		go func() {
			_, err := db.sendMessage(ctx, cLock, send, msg)
			ch <- err
		}()
	}

	var failed bool
	for i := 0; i < count; i++ {
		select {
		case err := <-ch:
			if err != nil {
				log.Log(ctx, fmt.Errorf("send message: %w", err))
				failed = true
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if failed {
		return errors.New("one or more failures when sending")
	}

	return nil
}

func (db *DB) sendMessage(ctx context.Context, cLock *processinglock.Conn, send SendFunc, m *Message) (bool, error) {
	ctx = log.WithFields(ctx, log.Fields{
		"DestTypeID": m.Dest.ID,
		"DestType":   m.Dest.DestType(),
		"CallbackID": m.ID,
	})

	if m.AlertID != 0 {
		ctx = log.WithField(ctx, "AlertID", m.AlertID)
	}
	_, err := cLock.Exec(ctx, db.setSending, m.ID)
	if err != nil {
		return false, err
	}
	sCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	var status *notification.SendResult
	err = retry.DoTemporaryError(func(int) error {
		status, err = send(sCtx, m)
		return err
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(65*time.Millisecond),
	)
	cancel()

	var pID notification.ProviderMessageID
	if status != nil {
		pID = status.ProviderMessageID
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

	if status.State == notification.StateFailedTemp {
		err = retryExec(db.tempFail, m.ID, pID, status.Details)
		return false, errors.Wrap(err, "mark failed message (temp)")
	}
	if status.State == notification.StateFailedPerm {
		err = retryExec(db.permFail, m.ID, pID, status.Details)
		return false, errors.Wrap(err, "mark failed message (perm)")
	}

	return true, errors.Wrap(db.UpdateMessageStatus(ctx, status), "update message status")
}
