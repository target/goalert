package calendarsubscription

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
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
}

// NewStore will create a new Store with the given parameters.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
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

// FindOne will return a single calendar subscription for the given id.
func (b *Store) FindOne(ctx context.Context, id string) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ID", id)
	if err != nil {
		return nil, err
	}

	var cs CalendarSubscription
	err = cs.scanFrom(b.findOne.QueryRowContext(ctx, id).Scan)
	if err == sql.ErrNoRows {
		return nil, validation.NewFieldError("ID", "not found")
	}
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// CreateTx will return a created calendar subscription with the given input.
func (b *Store) CreateTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(cs.UserID))
	if err != nil {
		return nil, err
	}

	n, err := cs.Normalize()
	if err != nil {
		return nil, err
	}

	cfgData, err := json.Marshal(n.Config)
	if err != nil {
		return nil, err
	}

	_, err = wrapTx(ctx, tx, b.create).ExecContext(ctx, n.ID, n.Name, n.UserID, n.Disabled, n.ScheduleID, cfgData)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// FindOneForUpdateTx will return a CalendarSubscription for the given userID that is locked for updating.
func (b *Store) FindOneForUpdateTx(ctx context.Context, tx *sql.Tx, userID, id string) (*CalendarSubscription, error) {
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
	row := wrapTx(ctx, tx, b.findOneUpd).QueryRowContext(ctx, id, userID)
	err = cs.scanFrom(row.Scan)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// UpdateTx updates a calendar subscription with given information.
func (b *Store) UpdateTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) error {
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

	_, err = wrapTx(ctx, tx, b.update).ExecContext(ctx, cs.ID, cs.UserID, cs.Name, cs.Disabled, cfgData)
	return err
}

// FindAllByUser returns all calendar subscriptions of a user.
func (b *Store) FindAllByUser(ctx context.Context, userID string) ([]CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	rows, err := b.findAll.QueryContext(ctx, userID)
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
func (b *Store) DeleteTx(ctx context.Context, tx *sql.Tx, userID string, ids ...string) error {
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

	_, err = wrapTx(ctx, tx, b.delete).ExecContext(ctx, sqlutil.UUIDArray(ids), userID)
	return err
}

// ICal ...
/*func ICal(shifts []oncall.Shift, start, end time.Time, valarm bool) (*os.File, error) {
	type iCalOptions struct {
		Shifts []oncall.Shift `json:"s,omitempty"`
		Start  time.Time      `json:"st,omitempty"`
		End    time.Time      `json:"e,omitempty"`
		Valarm bool           `jso n:"v, omitempty"`
	}

	// todo: valarm trigger time (always in minutes?)
	// todo: handle negative valarm values too
	// todo: fetch valarm from config column db

	var iCalTemplate = `
		BEGIN:VCALENDAR
		VERSION:2.0
		PRODID:-//ZContent.net//Zap Calendar 1.0//EN
		CALSCALE:GREGORIAN
		METHOD:PUBLISH
		{{range .Shifts}}
		BEGIN:VEVENT
		SUMMARY:On-Call
		DTSTART:{{.Start}}
		DTEND:{{.End}}
		END:VEVENT
		{{end}}

		{{if .Valarm}}
		BEGIN:VALARM
		ACTION:DISPLAY
		DESCRIPTION:REMINDER
		TRIGGER:-PT15M
		END:VALARM
		{{end}}

		END:VCALENDAR`

	iCal, err := template.New("iCal").Parse(iCalTemplate)
	if err != nil {

		return nil, err
	}

	i := iCalOptions{shifts, start, end, valarm}

	// Create output file
	file, err := os.Create("iCal.ics")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	err = iCal.Execute(buf, i)
	if err != nil {
		log.Fatal("render:", err)
	}

	// Write output to ics file
	err = ioutil.WriteFile("iCal.ics", buf.Bytes(), 0644)
	if err != nil {
		log.Fatal("save:", err)
	}
	return file, nil
}*/
