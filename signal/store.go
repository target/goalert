package signal

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateMany(ctx context.Context, signals []Signal) ([]*Signal, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer sqlutil.Rollback(ctx, "signal: create", tx)

	createdSignals := make([]*Signal, len(signals))
	for i := range signals {
		createdSignals[i], err = s.Create(ctx, tx, &signals[i])
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return createdSignals, nil
}

func (s *Store) Create(ctx context.Context, dbtx gadb.DBTX, sig *Signal) (*Signal, error) {
	serviceUUID, err2 := validate.ParseUUID("ServiceID", sig.ServiceID)
	ruleUUID, err1 := validate.ParseUUID("ServiceRuleID", sig.ServiceRuleID)

	err := validate.Many(err1, err2)
	if err != nil {
		return nil, err
	}

	rawPayload, err := json.Marshal(sig.OutgoingPayload)
	if err != nil {
		return nil, err
	}

	row, err := gadb.New(dbtx).SignalInsert(ctx, gadb.SignalInsertParams{
		ServiceRuleID:   ruleUUID,
		ServiceID:       serviceUUID,
		OutgoingPayload: rawPayload,
	})
	if err != nil {
		return nil, err
	}

	return &Signal{
		ID:              int(row.ID),
		ServiceID:       sig.ServiceID,
		ServiceRuleID:   sig.ServiceRuleID,
		OutgoingPayload: sig.OutgoingPayload,
		Timestamp:       row.Timestamp,
	}, nil
}
