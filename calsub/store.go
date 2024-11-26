package calsub

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Store allows the lookup and management of calendar subscriptions
type Store struct {
	db *sql.DB

	keys keyring.Keyring
	oc   *oncall.Store
}

// NewStore will create a new Store with the given parameters.
func NewStore(ctx context.Context, db *sql.DB, apiKeyring keyring.Keyring, oc *oncall.Store) (*Store, error) {
	return &Store{
		db:   db,
		keys: apiKeyring,
		oc:   oc,
	}, nil
}

// Authorize will return an authorized context associated with the given token. If the token is invalid
// or otherwise can not be authenticated, an error is returned.
func (s *Store) Authorize(ctx context.Context, tok authtoken.Token) (context.Context, error) {
	if tok.Type != authtoken.TypeCalSub {
		return ctx, permission.Unauthorized()
	}

	userID, err := gadb.NewCompat(s.db).CalSubAuthUser(ctx, gadb.CalSubAuthUserParams{
		ID:        tok.ID,
		CreatedAt: pgtype.Timestamptz{Time: tok.CreatedAt, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ctx, permission.Unauthorized()
	}
	if err != nil {
		return ctx, err
	}

	return permission.UserSourceContext(ctx, userID.String(), permission.RoleUser, &permission.SourceInfo{
		Type: permission.SourceTypeCalendarSubscription,
		ID:   tok.ID.String(),
	}), nil
}

// FindOne will return a single calendar subscription for the given id.
func (s *Store) FindOne(ctx context.Context, id string) (*Subscription, error) {
	return s._FindOne(ctx, gadb.NewCompat(s.db), id, false)
}

func (s *Store) _FindOne(ctx context.Context, q *gadb.Queries, id string, upd bool) (*Subscription, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ID", id)
	if err != nil {
		return nil, err
	}

	var sub gadb.FindOneCalSubRow

	if upd {
		uSub, uErr := q.FindOneCalSubForUpdate(ctx, uuid.MustParse(id))
		sub = gadb.FindOneCalSubRow(uSub)
		err = uErr
	} else {
		sub, err = q.FindOneCalSub(ctx, uuid.MustParse(id))
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, validation.NewFieldError("ID", "not found")
	}

	cs := Subscription{
		ID:         sub.ID.String(),
		Name:       sub.Name,
		UserID:     sub.UserID.String(),
		Disabled:   sub.Disabled,
		ScheduleID: sub.ScheduleID.String(),
		LastAccess: sub.LastAccess.Time,
	}
	err = json.Unmarshal(sub.Config, &cs.Config)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func (s *Store) FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Subscription, error) {
	return s._FindOne(ctx, gadb.NewCompat(tx), id, true)
}

// UpdateTx will update the given calendar subscription with the given input.
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, cs *Subscription) error {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(cs.UserID))
	if err != nil {
		return err
	}

	n, err := cs.Normalize()
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.Range("ReminderMinutes", len(n.Config.ReminderMinutes), 0, 15),
		validate.IDName("Name", n.Name),
		validate.UUID("ID", n.ID),
	)
	if err != nil {
		return err
	}

	cfgData, err := json.Marshal(n.Config)
	if err != nil {
		return err
	}

	err = gadb.NewCompat(tx).UpdateCalSub(ctx, gadb.UpdateCalSubParams{
		ID:       uuid.MustParse(n.ID),
		Name:     n.Name,
		Disabled: n.Disabled,
		Config:   cfgData,
		UserID:   uuid.MustParse(n.UserID),
	})
	return err
}

// CreateTx will return a created calendar subscription with the given input.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, cs *Subscription) (*Subscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(cs.UserID))
	if err != nil {
		return nil, err
	}

	cfg := config.FromContext(ctx)
	if cfg.General.DisableCalendarSubscriptions {
		return nil, validation.NewGenericError("disabled by administrator")
	}

	n, err := cs.Normalize()
	if err != nil {
		return nil, err
	}

	cfgData, err := json.Marshal(n.Config)
	if err != nil {
		return nil, err
	}

	now, err := gadb.NewCompat(tx).CreateCalSub(ctx, gadb.CreateCalSubParams{
		ID:         uuid.MustParse(n.ID),
		Name:       n.Name,
		UserID:     uuid.MustParse(n.UserID),
		Disabled:   n.Disabled,
		ScheduleID: uuid.MustParse(n.ScheduleID),
		Config:     cfgData,
	})
	if err != nil {
		return nil, err
	}

	tokID, err := uuid.Parse(n.ID)
	if err != nil {
		return nil, err
	}

	n.token, err = authtoken.Token{
		Type:      authtoken.TypeCalSub,
		Version:   2,
		CreatedAt: now.Time,
		ID:        tokID,
	}.Encode(s.keys.Sign)
	return n, err
}

// FindAllByUser returns all calendar subscriptions of a user.
func (s *Store) FindAllByUser(ctx context.Context, userID string) ([]Subscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	subs, err := gadb.NewCompat(s.db).FindManyCalSubByUser(ctx, uuid.MustParse(userID))
	if err != nil {
		return nil, err
	}

	cs := make([]Subscription, len(subs))
	for i, sub := range subs {
		cs[i] = Subscription{
			ID:         sub.ID.String(),
			Name:       sub.Name,
			UserID:     sub.UserID.String(),
			Disabled:   sub.Disabled,
			ScheduleID: sub.ScheduleID.String(),
			LastAccess: sub.LastAccess.Time,
		}
		err = json.Unmarshal(sub.Config, &cs[i].Config)
		if err != nil {
			return nil, err
		}
	}

	return cs, nil
}

// DeleteTx removes calendar subscriptions with the given ids for the given user.
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, userID string, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.ManyUUID("ID", ids, 50),
		validate.UUID("UserID", userID),
	)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	uids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uids[i] = uuid.MustParse(id)
	}

	return gadb.NewCompat(tx).DeleteManyCalSub(ctx, gadb.DeleteManyCalSubParams{
		Column1: uids,
		UserID:  uuid.MustParse(userID),
	})
}
