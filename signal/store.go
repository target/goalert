package signal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

const maxBatch = 500

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) FindOne(ctx context.Context, id int) (*Signal, error) {
	signals, err := s.FindMany(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	// If signal is not found
	if len(signals) == 0 {
		return nil, sql.ErrNoRows
	}
	return &signals[0], nil
}

func (s *Store) FindMany(ctx context.Context, signalIDs []int) ([]Signal, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}
	if len(signalIDs) == 0 {
		return nil, nil
	}

	err = validate.Range("SignalIDs", len(signalIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(signalIDs))
	for _, id := range signalIDs {
		ids = append(ids, int64(id))
	}

	rows, err := gadb.New(s.db).SignalFindMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	signals := make([]Signal, 0, len(signalIDs))

	for _, row := range rows {
		payload := make(map[string]interface{})
		err := json.Unmarshal(row.OutgoingPayload, &payload)
		if err != nil {
			log.Log(log.WithField(ctx, "SignalID", row.ID), fmt.Errorf("unmarshal signal payload: %w", err))
		}
		signals = append(signals, Signal{
			ID:              row.ID,
			ServiceID:       row.ServiceID.String(),
			ServiceRuleID:   row.ServiceRuleID.String(),
			OutgoingPayload: payload,
			Scheduled:       row.Scheduled,
			Timestamp:       row.Timestamp,
		})
	}

	return signals, nil
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
		ID:              row.ID,
		ServiceID:       sig.ServiceID,
		ServiceRuleID:   sig.ServiceRuleID,
		OutgoingPayload: sig.OutgoingPayload,
		Timestamp:       row.Timestamp,
	}, nil
}
