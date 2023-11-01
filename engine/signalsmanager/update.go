package signalsmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
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

type content struct {
	prop  string
	value string
}

type destination struct {
	DestinationType string    `json:"dest_type"`
	DestinationID   string    `json:"dest_id"`
	DestinationVal  string    `json:"dest_value"`
	Content         []content `json:"contents"`
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

	srvRule, err := q.SvcRuleFindOne(ctx, sig.ServiceRuleID)

	destList := []destination{}
	err = json.Unmarshal(srvRule.Actions.RawMessage, &destList)

	for _, dest := range destList {
		err = q.SignalsManagerSendOutgoing(ctx, gadb.SignalsManagerSendOutgoingParams{
			SignalID:        int32(sig.ID),
			ServiceID:       sig.ServiceID,
			DestinationType: dest.DestinationType,
			DestinationID:   dest.DestinationID,
			DestinationVal:  dest.DestinationVal,
			Content:         srvRule.Actions.RawMessage,
		})
		if err != nil {
			return fmt.Errorf("insert outgoing_signals error: %w", err)
		}
	}

	return q.SignalsManagerSetScheduled(ctx, sig.ID)
}
