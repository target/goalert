package signal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user/contactmethod"
)

const engineVersion = 1

// DB handles updating metrics
type DB struct {
	lock *processinglock.Lock
	db   *sql.DB
}

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: engineVersion,
		Type:    processinglock.TypeSignals,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock: lock,
		db:   db,
	}, nil
}

// SendFunc defines a function that sends signals.
type SendFunc func(context.Context, *OutgoingSignal) (*notification.SendResult, error)

// SendMessages will send notifications using SendFunc.
func (db *DB) SendSignals(ctx context.Context, send SendFunc) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		err = db._sendSignal(ctx, send, db.db)
		if errors.Is(err, errDone) {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

var errDone = errors.New("done")

func (db *DB) _sendSignal(ctx context.Context, send SendFunc, tx gadb.DBTX) error {
	q := gadb.New(tx)

	sig, err := q.OutgoingSignalFindNext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return errDone
	}
	if err != nil {
		return fmt.Errorf("find next signal error: %w", err)
	}

	// TODO: refine, "SLACK" overlaps both nc and cm.
	var sigType notification.ScannableDestType
	sigType.NC = notificationchannel.Type(sig.DestinationType)
	sigType.CM = contactmethod.Type(sig.DestinationType)

	signal := OutgoingSignal{
		ID:   sig.ID.String(),
		Type: notification.MessageTypeSignal,
		Dest: notification.Dest{
			ID:    sig.DestinationID,
			Type:  sigType.DestType(),
			Value: sig.DestinationVal,
		},
		SignalID:  int(sig.SignalID),
		UserID:    permission.UserID(ctx),
		ServiceID: sig.ServiceID.String(),

		CreatedAt: time.Now(),
		SentAt:    time.Now(),

		Message: sig.Message,
		Content: sig.Content,
	}

	res, err := send(ctx, &signal)
	if err != nil {
		return fmt.Errorf("send signal error: %w", err)
	}

	if res.Status.State.IsOK() {
		err = q.OutgoingSignalUpdateSent(ctx, sig.ID)
	}
	if err != nil {
		return fmt.Errorf("outgoing signal status error: %w", err)
	}

	return nil
}
