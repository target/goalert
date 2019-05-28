package message

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// DB implements a priority message sender using Postgres.
type DB struct {
	lock *processinglock.Lock

	c Config

	stuckMessages *sql.Stmt

	setSending *sql.Stmt

	lockStmt    *sql.Stmt
	pending     *sql.Stmt
	currentTime *sql.Stmt
	retryReset  *sql.Stmt
	retryClear  *sql.Stmt

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
		pending: p.P(fmt.Sprintf(`
			select
				msg.id,
				msg.message_type,
				msg.contact_method_id,
				cm.type,
				msg.alert_id,
				msg.alert_log_id,
				msg.user_verification_code_id,
				msg.channel_id,
				chan.type
			from outgoing_messages msg
			left join user_contact_methods cm on cm.id = msg.contact_method_id
			left join notification_channels chan on chan.id = msg.channel_id
			where last_status = 'pending' and (not cm isnull or not chan isnull)
			order by
				msg.message_type,
				(select max(sent_at) from outgoing_messages om where om.escalation_policy_id = msg.escalation_policy_id) nulls first,
				(select max(sent_at) from outgoing_messages om where om.service_id = msg.service_id) nulls first,
				(select max(sent_at) from outgoing_messages om where om.alert_id = msg.alert_id) nulls first,
				channel_id,
				(select max(sent_at) from outgoing_messages om where om.user_id = msg.user_id) nulls first,
				(select max(sent_at) from outgoing_messages om where om.contact_method_id = msg.contact_method_id) nulls first,
				msg.created_at,
				msg.alert_id,
				msg.alert_log_id,
				msg.contact_method_id
			limit %d
		`, c.MaxMessagesPerCycle)),
	}, p.Err
}

func (db *DB) getRows(ctx context.Context, tx *sql.Tx) ([]row, error) {
	rows, err := tx.Stmt(db.pending).QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch outgoing messages")
	}
	defer rows.Close()

	var result []row
	var r row
	for rows.Next() {
		var cmID, cmType, chID, chType sql.NullString
		err = rows.Scan(
			&r.ID,
			&r.Type,
			&cmID,
			&cmType,
			&r.AlertID,
			&r.AlertLogID,
			&r.VerifyID,
			&chID,
			&chType,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan row")
		}
		switch {
		case cmType.String == string(contactmethod.TypeSMS):
			r.DestType = notification.DestTypeSMS
			r.DestID = cmID.String
		case cmType.String == string(contactmethod.TypeVoice):
			r.DestType = notification.DestTypeVoice
			r.DestID = cmID.String
		case chType.String == string(notificationchannel.TypeSlack):
			r.DestType = notification.DestTypeSlackChannel
			r.DestID = chID.String
		default:
			log.Debugf(ctx, "unknown message type for message %s", r.ID)
			continue
		}

		result = append(result, r)
	}

	return result, nil
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
	rows, _ := res.RowsAffected()
	if rows > 0 {
		log.Log(execCtx, errors.Errorf("terminated %d stale backend instance(s) holding message sending lock", rows))
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

	_, err = tx.Stmt(db.failDisabledCM).ExecContext(execCtx)
	if err != nil {
		return errors.Wrap(err, "check for disabled CMs")
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

	msgs, err := db.getRows(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "get pending messages")
	}

	counts := make(batchCounts, len(db.c.RateLimit))
	for cType, cfg := range db.c.RateLimit {
		if cfg == nil || cfg.Batch <= 0 || cfg.PerSecond < 1 || !cType.IsUserCM() {
			continue
		}
		var c int
		err = tx.Stmt(db.sentByCMType).QueryRowContext(execCtx, t.Add(-cfg.Batch), contactmethod.TypeFromDestType(cType)).Scan(&c)
		if err != nil {
			return errors.Wrap(err, "get sent message count")
		}
		counts[cType] = c
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit message updates")
	}

	if len(msgs) > 0 {
		msgByType := make(map[notification.DestType][]row)

		for _, m := range msgs {
			msgByType[m.DestType] = append(msgByType[m.DestType], m)
		}

		// ensure we cancel sending other messages on err
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// ensure the buffer is large enough to hold all responses, even if we exit on err
		// otherwise the goroutine will hang and be a memory leak
		errCh := make(chan error, len(msgByType))

		for typ, rows := range msgByType {
			toSend := db.c.batchNum(typ) // max messages per cycle
			if toSend == 0 {
				// no limit
				go db.sendAllMessages(ctx, cLock, send, rows, 0, errCh)
				continue
			}

			toSend -= counts[typ]
			if toSend <= 0 {
				// nothing to send
				errCh <- nil
				continue
			}

			// only send remaining in queue
			go db.sendAllMessages(ctx, cLock, send, rows, toSend, errCh)
		}

		n := 0
		for err := range errCh {
			n++
			if err != nil && errors.Cause(err) != processinglock.ErrNoLock {
				log.Log(ctx, errors.Wrap(err, "send"))
			}
			// jump out once we've completed all types
			if n == len(msgByType) {
				break
			}
		}
	}

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
			State:             notification.MessageStateActive,
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

func (db *DB) sendAllMessages(ctx context.Context, cLock *processinglock.Conn, send SendFunc, rows []row, count int, errCh chan error) {
	type sendResult struct {
		sent bool
		err  error
	}

	ch := make(chan sendResult, len(rows)) // ensure we can store all responses if needed
	doSend := func(r row) {
		var res sendResult
		res.sent, res.err = db.sendMessage(ctx, cLock, send, &r)
		ch <- res
	}

	var sent int
	var pending int
	for i, m := range rows {
		if db.c.Pausable.IsPausing() {
			// abort due to pause
			break
		}
		go doSend(m)
		pending++
		if i < 20 && (i < count || count == 0) {
			continue
		}

		res := <-ch
		pending--
		if res.err != nil {
			errCh <- res.err
			return
		}
		if res.sent {
			sent++
		}
		if count > 0 && sent == count {
			break
		}
	}

	for ; pending > 0; pending-- {
		// check remaining responses for errors
		res := <-ch
		if res.err != nil {
			errCh <- res.err
			return
		}
	}

	errCh <- nil
}

func (db *DB) sendMessage(ctx context.Context, cLock *processinglock.Conn, send SendFunc, m *row) (bool, error) {
	ctx, sp := trace.StartSpan(ctx, "Engine.MessageManager.SendMessage")
	defer sp.End()
	ctx = log.WithFields(ctx, log.Fields{
		"DestTypeID": m.DestID,
		"DestType":   m.DestType.String(),
		"CallbackID": m.ID,
	})
	sp.AddAttributes(
		trace.StringAttribute("message.dest.id", m.DestID),
		trace.StringAttribute("message.dest.type", m.DestType.String()),
		trace.StringAttribute("message.callback.id", m.ID),
	)
	var alertID int
	if m.AlertID.Valid {
		alertID = int(m.AlertID.Int64)
		ctx = log.WithField(ctx, "AlertID", alertID)
	}
	_, err := cLock.Exec(ctx, db.setSending, m.ID)
	if err != nil {
		return false, err
	}
	sCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	var status *notification.MessageStatus
	err = retry.DoTemporaryError(func(int) error {
		status, err = send(sCtx, &Message{
			ID:         m.ID,
			Type:       m.Type,
			DestType:   m.DestType,
			DestID:     m.DestID,
			AlertID:    alertID,
			AlertLogID: int(m.AlertLogID.Int64),
			VerifyID:   m.VerifyID.String,
		})
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
