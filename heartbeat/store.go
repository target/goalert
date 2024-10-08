package heartbeat

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// Store manages heartbeat checks and recording heartbeats.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store and prepares all sql statements.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	return &Store{
		db: db,
	}, nil
}

// CreateTx creates a new heartbeat Monitor.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, m *Monitor) (*Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	n, err := m.Normalize()
	if err != nil {
		return nil, err
	}
	id := uuid.New()
	n.ID = id.String()

	err = s.dbtx(tx).HBInsert(ctx, gadb.HBInsertParams{
		ID:                id,
		Name:              n.Name,
		ServiceID:         uuid.MustParse(n.ServiceID), // already validated in Normalize
		HeartbeatInterval: sqlutil.IntervalMicro(n.Timeout),
		AdditionalDetails: sql.NullString{String: n.AdditionalDetails, Valid: n.AdditionalDetails != ""},
	})
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (s *Store) dbtx(tx *sql.Tx) *gadb.Queries {
	db := gadb.New(s.db)
	if tx == nil {
		return db
	}

	return db.WithTx(tx)
}

// RecordHeartbeat records a heartbeat for the given heartbeat ID.
func (s *Store) RecordHeartbeat(ctx context.Context, idStr string) error {
	id, err := validate.ParseUUID("MonitorID", idStr)
	if err != nil {
		return err
	}

	return s.dbtx(nil).HBRecordHeartbeat(ctx, id)
}

// DeleteTx deletes the heartbeat check with the given ID(s).
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, idStrs ...string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}

	ids, err := validate.ParseManyUUID("MonitorID", idStrs, 100)
	if err != nil {
		return err
	}

	return s.dbtx(tx).HBDelete(ctx, ids)
}

// UpdateTx updates a heartbeat Monitor.
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, m *Monitor) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}

	n, err := m.Normalize()
	if err != nil {
		return err
	}

	id, err := validate.ParseUUID("MonitorID", n.ID)
	if err != nil {
		return err
	}

	return s.dbtx(tx).HBUpdate(ctx, gadb.HBUpdateParams{
		ID:                id,
		Name:              n.Name,
		HeartbeatInterval: sqlutil.IntervalMicro(n.Timeout),
		AdditionalDetails: sql.NullString{String: n.AdditionalDetails, Valid: n.AdditionalDetails != ""},
		Muted:             sql.NullString{String: n.Muted, Valid: n.Muted != ""},
	})
}

// FindOneTx returns a heartbeat montior for updating.
func (s *Store) FindOneTx(ctx context.Context, tx *sql.Tx, idStr string) (*Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	id, err := validate.ParseUUID("ID", idStr)
	if err != nil {
		return nil, err
	}

	m, err := s.dbtx(tx).HBByIDForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}

	mon := fromDB(m)
	return &mon, nil
}

// FindMany returns the heartbeat monitors with the given IDs.
//
// The order and number of returned monitors is not guaranteed.
func (s *Store) FindMany(ctx context.Context, idStrs []string) ([]Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	if len(idStrs) == 0 {
		return nil, nil
	}

	ids, err := validate.ParseManyUUID("IDs", idStrs, search.MaxResults)
	if err != nil {
		return nil, err
	}

	rows, err := s.dbtx(nil).HBManyByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make([]Monitor, 0, len(rows))
	for _, m := range rows {
		res = append(res, fromDB(m))
	}

	return res, nil
}

func fromDB(m gadb.HeartbeatMonitor) Monitor {
	return Monitor{
		ID:                m.ID.String(),
		Name:              m.Name,
		ServiceID:         m.ServiceID.String(),
		Timeout:           time.Duration(m.HeartbeatInterval.Microseconds) * time.Microsecond,
		AdditionalDetails: m.AdditionalDetails.String,
		Muted:             m.Muted.String,
		lastState:         State(m.LastState),
		lastHeartbeat:     m.LastHeartbeat.Time,
	}
}

// FindAllByService returns all heartbeats belonging to the given service ID.
func (s *Store) FindAllByService(ctx context.Context, serviceIDStr string) ([]Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	serviceID, err := validate.ParseUUID("ServiceID", serviceIDStr)
	if err != nil {
		return nil, err
	}

	rows, err := s.dbtx(nil).HBByService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	res := make([]Monitor, 0, len(rows))
	for _, m := range rows {
		res = append(res, fromDB(m))
	}

	return res, nil
}
