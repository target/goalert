package signalsmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/target/goalert/engine/signal"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/service/rule"
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
	if err != nil {
		return fmt.Errorf("find signal service rule error: %w", err)
	}

	actionList := []rule.Action{}
	err = json.Unmarshal(srvRule.Actions.RawMessage, &actionList)
	if err != nil {
		return fmt.Errorf("unmarshal service rule actions error: %w", err)
	}

	for _, action := range actionList {
		// aquires information nested in the service rule's actions
		// messages could use processing for variables
		// destination value may be sent as a Content object inside actions
		// content is reorganized to include only necessary aditional context
		message, destVal, content, err := signal.ProcessContent(action)
		if err != nil {
			return fmt.Errorf("signalsmanager process signal content error: %w", err)
		}

		err = q.SignalsManagerSendOutgoing(ctx, gadb.SignalsManagerSendOutgoingParams{
			SignalID:        int32(sig.ID),
			ServiceID:       sig.ServiceID,
			DestinationType: action.DestType,
			DestinationVal:  destVal,
			Message:         message,
			Content:         content,
		})
		if err != nil {
			return fmt.Errorf("insert outgoing_signals error: %w", err)
		}
	}

	return q.SignalsManagerSetScheduled(ctx, sig.ID)
}
