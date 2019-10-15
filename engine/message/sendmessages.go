package message

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"
	"go.opencensus.io/trace"
)

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

	for {
		if sentCount() < toSend {
			msg := q.NextByType(typ)
			if msg == nil {
				break
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
