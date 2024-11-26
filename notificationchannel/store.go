package notificationchannel

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/gadb/pgxdb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db  *pgxdb.Queries
	reg *nfydest.Registry
}

func NewStore(ctx context.Context, db *pgxdb.Queries, reg *nfydest.Registry) (*Store, error) {
	return &Store{
		db:  db,
		reg: reg,
	}, nil
}

func (s *Store) FindMany(ctx context.Context, ids []string) ([]Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	uuids, err := validate.ParseManyUUID("ID", ids, search.MaxResults)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.NotifChanFindMany(ctx, uuids)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	channels := make([]Channel, len(rows))
	for i, r := range rows {
		channels[i].fromRow(r)
	}

	return channels, nil
}

func (s *Store) FindDestByID(ctx context.Context, id uuid.UUID) (gadb.DestV1, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return gadb.DestV1{}, err
	}

	row, err := s.db.NotifChanFindOne(ctx, id)
	if err != nil {
		return gadb.DestV1{}, err
	}

	return row.Dest.DestV1, nil
}

func (s *Store) LookupDestID(ctx context.Context, tx *sql.Tx, d gadb.DestV1) (uuid.UUID, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return uuid.UUID{}, err
	}

	return gadb.New(tx).NotifChanFindDestID(ctx, gadb.NullDestV1{Valid: true, DestV1: d})
}

func (s *Store) MapDestToID(ctx context.Context, tx gadb.DBTX, d gadb.DestV1) (uuid.UUID, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return uuid.UUID{}, err
	}
	err = s.reg.ValidateDest(ctx, d)
	if err != nil {
		return uuid.UUID{}, err
	}
	info, err := s.reg.DisplayInfo(ctx, d)
	if err != nil {
		return uuid.UUID{}, err
	}

	return gadb.New(tx).NotifChanUpsertDest(ctx, gadb.NotifChanUpsertDestParams{
		ID:   uuid.New(),
		Dest: gadb.NullDestV1{Valid: true, DestV1: d},
		Name: info.Text,
	})
}

func (s *Store) FindOne(ctx context.Context, id uuid.UUID) (*Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	var c Channel
	row, err := s.db.NotifChanFindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	c.fromRow(row)

	return &c, nil
}
