package processinglock

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util/jsonutil"
)

// State manages the state value for a processing lock.
type State struct {
	data json.RawMessage
	tx   *sql.Tx
	l    *Lock
}

// BeginTxWithState will start a transaction, returning a State object.
func (l *Lock) BeginTxWithState(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, *State, error) {
	tx, err := l.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	return tx, &State{tx: tx, l: l}, nil
}

// Load will load the JSON state from the database.
func (s *State) Load(ctx context.Context, v interface{}) (err error) {
	s.data, err = gadb.NewCompat(s.tx).ProcLoadState(ctx, gadb.EngineProcessingType(s.l.cfg.Type))
	if err != nil {
		return err
	}

	return json.Unmarshal(s.data, v)
}

// Save will save the JSON state to the database, taking care to ensure that
// existing unknown fields are preserved.
func (s *State) Save(ctx context.Context, v interface{}) error {
	data, err := jsonutil.Apply(s.data, v)
	if err != nil {
		return err
	}
	s.data = data

	err = gadb.NewCompat(s.tx).ProcSaveState(ctx, gadb.ProcSaveStateParams{
		TypeID: gadb.EngineProcessingType(s.l.cfg.Type),
		State:  s.data,
	})
	if err != nil {
		return err
	}

	return nil
}
