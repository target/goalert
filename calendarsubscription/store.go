package calendarsubscription

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/target/goalert/config"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Store allows the lookup and management of calendar subscriptions
type Store struct {
	db         *sql.DB
	findOne    *sql.Stmt
	create     *sql.Stmt
	update     *sql.Stmt
	delete     *sql.Stmt
	findAll    *sql.Stmt
	findOneUpd *sql.Stmt
	authUser   *sql.Stmt
	now        *sql.Stmt

	keys keyring.Keyring
	oc   oncall.Store
}

const tokenAudience = "ga-cal-sub"

// NewStore will create a new Store with the given parameters.
func NewStore(ctx context.Context, db *sql.DB, apiKeyring keyring.Keyring, oc oncall.Store) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db:   db,
		keys: apiKeyring,
		oc:   oc,

		now: p.P(`SELECT now()`),
		authUser: p.P(`
			UPDATE user_calendar_subscriptions
			SET last_access = now()
			WHERE NOT disabled AND id = $1 AND date_trunc('second', created_at) = $2
			RETURNING user_id
		`),
		findOne: p.P(`
			SELECT
				id, name, user_id, disabled, schedule_id, config, last_access
			FROM user_calendar_subscriptions
			WHERE id = $1
		`),
		create: p.P(`
			INSERT INTO user_calendar_subscriptions (
				id, name, user_id, disabled, schedule_id, config
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING created_at
		`),
		update: p.P(`
			UPDATE user_calendar_subscriptions
			SET name = $3, disabled = $4, config = $5, last_update = now()
			WHERE id = $1 AND user_id = $2
		`),
		delete: p.P(`
			DELETE FROM user_calendar_subscriptions
			WHERE id = any($1) AND user_id = $2
		`),
		findAll: p.P(`
			SELECT
				id, name, user_id, disabled, schedule_id, config, last_access
			FROM user_calendar_subscriptions
			WHERE user_id = $1
		`),
		findOneUpd: p.P(`
			SELECT
				id, name, user_id, disabled, schedule_id, config, last_access
			FROM user_calendar_subscriptions
			WHERE id = $1 AND user_id = $2
		`),
	}, p.Err
}

func isCreationDisabled (ctx context.Context) bool {
	cfg := config.FromContext(ctx)
	return cfg.General.DisableCalendarSubscriptions
}

func wrapTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}
	return tx.StmtContext(ctx, stmt)
}

func (cs *CalendarSubscription) scanFrom(scanFn func(...interface{}) error) error {
	var lastAccess sql.NullTime
	var cfgData []byte
	err := scanFn(&cs.ID, &cs.Name, &cs.UserID, &cs.Disabled, &cs.ScheduleID, &cfgData, &lastAccess)
	if err != nil {
		return err
	}

	cs.LastAccess = lastAccess.Time
	err = json.Unmarshal(cfgData, &cs.Config)
	return err
}

// Authorize will return an authorized context associated with the given token. If the token is invalid
// or otherwise can not be authenticated, an error is returned.
func (s *Store) Authorize(ctx context.Context, token string) (context.Context, error) {
	var c jwt.StandardClaims
	_, err := s.keys.VerifyJWT(token, &c)
	if err != nil {
		log.Debug(ctx, err)
		return ctx, validation.NewFieldError("token", "verification failed")
	}

	if !c.VerifyAudience(tokenAudience, true) {
		return ctx, validation.NewFieldError("aud", "invalid audience")
	}

	err = validate.UUID("sub", c.Subject)
	if err != nil {
		return ctx, err
	}

	var userID string
	err = s.authUser.QueryRowContext(ctx, c.Subject, time.Unix(c.IssuedAt, 0)).Scan(&userID)
	if err == sql.ErrNoRows {
		return ctx, validation.NewFieldError("sub", "invalid")
	}
	if err != nil {
		return ctx, err
	}

	return permission.UserSourceContext(ctx, userID, permission.RoleUser, &permission.SourceInfo{
		Type: permission.SourceTypeCalendarSubscription,
		ID:   c.Subject,
	}), nil
}

// FindOne will return a single calendar subscription for the given id.
func (s *Store) FindOne(ctx context.Context, id string) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ID", id)
	if err != nil {
		return nil, err
	}

	var cs CalendarSubscription
	err = cs.scanFrom(s.findOne.QueryRowContext(ctx, id).Scan)
	if err == sql.ErrNoRows {
		return nil, validation.NewFieldError("ID", "not found")
	}
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// CreateTx will return a created calendar subscription with the given input.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(cs.UserID))
	if err != nil {
		return nil, err
	}

	if isCreationDisabled(ctx) {
		return nil, validation.NewGenericError("creation disabled by administrator")
	}

	n, err := cs.Normalize()
	if err != nil {
		return nil, err
	}

	cfgData, err := json.Marshal(n.Config)
	if err != nil {
		return nil, err
	}

	var now time.Time
	row := wrapTx(ctx, tx, s.create).QueryRowContext(ctx, n.ID, n.Name, n.UserID, n.Disabled, n.ScheduleID, cfgData)
	err = row.Scan(&now)
	if err != nil {
		return nil, err
	}

	n.token, err = s.keys.SignJWT(jwt.StandardClaims{
		Subject:  n.ID,
		Audience: tokenAudience,
		IssuedAt: now.Unix(),
	})
	return n, err
}

// FindOneForUpdateTx will return a CalendarSubscription for the given userID that is locked for updating.
func (s *Store) FindOneForUpdateTx(ctx context.Context, tx *sql.Tx, userID, id string) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}
	err = validate.Many(
		validate.UUID("ID", id),
		validate.UUID("UserID", userID),
	)
	if err != nil {
		return nil, err
	}

	var cs CalendarSubscription
	row := wrapTx(ctx, tx, s.findOneUpd).QueryRowContext(ctx, id, userID)
	err = cs.scanFrom(row.Scan)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// UpdateTx updates a calendar subscription with given information.
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) error {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(cs.UserID))
	if err != nil {
		return err
	}

	n, err := cs.Normalize()
	if err != nil {
		return err
	}

	cfgData, err := json.Marshal(n.Config)
	if err != nil {
		return err
	}

	_, err = wrapTx(ctx, tx, s.update).ExecContext(ctx, cs.ID, cs.UserID, cs.Name, cs.Disabled, cfgData)
	return err
}

// FindAllByUser returns all calendar subscriptions of a user.
func (s *Store) FindAllByUser(ctx context.Context, userID string) ([]CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAll.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calendarsubscriptions []CalendarSubscription
	for rows.Next() {
		var cs CalendarSubscription
		err = cs.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}

		calendarsubscriptions = append(calendarsubscriptions, cs)
	}

	return calendarsubscriptions, nil
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

	_, err = wrapTx(ctx, tx, s.delete).ExecContext(ctx, sqlutil.UUIDArray(ids), userID)
	return err
}
