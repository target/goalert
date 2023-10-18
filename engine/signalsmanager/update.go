package signalsmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	for i := 0; i < 50; i++ {
		err = db.lock.WithTx(ctx, db.update)
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

type payload struct {
	Destination string          `json:"destination_id"`
	Message     json.RawMessage `json:"received_payload"`
}

func (db *DB) update(ctx context.Context, tx *sql.Tx) error {
	q := gadb.New(tx)

	// gets the next signal in db that is not locked
	sig, err := q.SignalsManagerFindNext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return errDone
	}
	if err != nil {
		return fmt.Errorf("find next signal error: %w", err)
	}

	pay := payload{}
	err = json.Unmarshal(sig.OutgoingPayload, &pay)
	if err != nil {
		log.Log(log.WithField(ctx, "SignalID", sig.ID), fmt.Errorf("unmarshal signal payload: %w", err))
		return err
	}

	err = q.SignalsManagerSendOutgoing(ctx, gadb.SignalsManagerSendOutgoingParams{
		ServiceID:       sig.ServiceID,
		OutgoingPayload: sig.OutgoingPayload,
		ChannelID:       uuid.MustParse(pay.Destination),
	})
	if err != nil {
		return fmt.Errorf("insert outgoing_signals error: %w", err)
	}

	return q.SignalsManagerSetScheduled(ctx, sig.ID)
}
