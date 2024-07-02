package notificationchannel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
	}, p.Err
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

	rows, err := gadb.New(s.db).NotifChanFindMany(ctx, uuids)
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

func (s *Store) MapToID(ctx context.Context, tx *sql.Tx, c *Channel) (uuid.UUID, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return uuid.UUID{}, err
	}

	n, err := c.Normalize()
	if err != nil {
		return uuid.UUID{}, err
	}

	db := gadb.New(s.db)
	if tx != nil {
		db = db.WithTx(tx)
	}

	row, err := db.NotifChanFindByValue(ctx, gadb.NotifChanFindByValueParams{
		Type:  gadb.EnumNotifChannelType(n.Type),
		Value: n.Value,
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("lookup existing entry: %w", err)
	}

	if row.Name == c.Name {
		// short-circuit if it already exists and is up-to-date.
		return row.ID, nil
	}

	var ownTx bool
	if tx == nil {
		ownTx = true
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("start tx: %w", err)
		}
		defer sqlutil.Rollback(ctx, "notificationchannel: map channel ID to UUID", tx)
		db = db.WithTx(tx)
	}

	err = db.NotifChanLock(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("acquire lock: %w", err)
	}

	// try again after exclusive lock
	row, err = db.NotifChanFindByValue(ctx, gadb.NotifChanFindByValueParams{
		Type:  gadb.EnumNotifChannelType(n.Type),
		Value: n.Value,
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("lookup existing entry exclusively: %w", err)
	}
	if row.Name == c.Name {
		// short-circuit if it already exists and is up-to-date.
		return row.ID, nil
	}

	if row.ID == uuid.Nil {
		// create new one
		row.ID = uuid.New()
		err = db.NotifChanCreate(ctx, gadb.NotifChanCreateParams{
			ID:    row.ID,
			Name:  n.Name,
			Type:  gadb.EnumNotifChannelType(n.Type),
			Value: n.Value,
		})
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("create new NC: %w", err)
		}
	} else {
		// update existing name
		err = db.NotifChanUpdateName(ctx, gadb.NotifChanUpdateNameParams{
			ID:   row.ID,
			Name: n.Name,
		})
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("update NC name: %w", err)
		}
	}

	if ownTx {
		err = tx.Commit()
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("commit tx: %w", err)
		}
	}

	return row.ID, nil
}

func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return err
	}

	uuids, err := validate.ParseManyUUID("ID", ids, 100)
	if err != nil {
		return err
	}

	db := gadb.New(s.db)
	if tx != nil {
		db = db.WithTx(tx)
	}

	return db.NotifChanDeleteMany(ctx, uuids)
}

func (s *Store) FindOne(ctx context.Context, id uuid.UUID) (*Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	var c Channel
	row, err := gadb.New(s.db).NotifChanFindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	c.fromRow(row)

	return &c, nil
}
