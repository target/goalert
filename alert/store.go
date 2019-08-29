package alert

import (
	"context"
	"database/sql"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const maxBatch = 500

type Store interface {
	Manager
	Create(context.Context, *Alert) (*Alert, error)

	// CreateOrUpdate will create an alert or log a "duplicate suppressed message" if
	// Status is Triggered. If Status is Closed, it will close and return the result.
	//
	// In the case that Status is closed but a matching alert is not present, nil is returned.
	// Otherwise the current alert is returned.
	CreateOrUpdate(context.Context, *Alert) (*Alert, error)

	// CreateOrUpdateTx returns `isNew` to indicate if the returned alert was a new one.
	// It is the caller's responsibility to log alert creation if the transaction is committed (and isNew is true).
	CreateOrUpdateTx(context.Context, *sql.Tx, *Alert) (a *Alert, isNew bool, err error)

	FindAllSummary(ctx context.Context) ([]Summary, error)
	Escalate(ctx context.Context, alertID int, currentLevel int) error
	EscalateMany(ctx context.Context, alertIDs []int) ([]int, error)
	GetCreationTime(ctx context.Context, alertID int) (time.Time, error)

	LegacySearch(ctx context.Context, opt *LegacySearchOptions) ([]Alert, int, error)
	Search(ctx context.Context, opts *SearchOptions) ([]Alert, error)
	State(ctx context.Context, alertIDs []int) ([]State, error)
}
type Manager interface {
	FindOne(context.Context, int) (*Alert, error)
	FindMany(context.Context, []int) ([]Alert, error)
	UpdateStatus(context.Context, int, Status) error
	UpdateStatusByService(ctx context.Context, serviceID string, status Status) error
	UpdateManyAlertStatus(ctx context.Context, status Status, alertIDs []int) (updatedAlertIDs []int, err error)
	UpdateStatusTx(context.Context, *sql.Tx, int, Status) error
	EPID(ctx context.Context, alertID int) (string, error)
}

type DB struct {
	db    *sql.DB
	logDB alertlog.Store

	insert          *sql.Stmt
	update          *sql.Stmt
	logs            *sql.Stmt
	findAllSummary  *sql.Stmt
	findMany        *sql.Stmt
	getCreationTime *sql.Stmt
	getServiceID    *sql.Stmt

	lockSvc      *sql.Stmt
	lockAlertSvc *sql.Stmt

	getStatusAndLockSvc *sql.Stmt

	createUpdNew   *sql.Stmt
	createUpdAck   *sql.Stmt
	createUpdClose *sql.Stmt

	updateByStatusAndService *sql.Stmt
	updateByIDAndStatus      *sql.Stmt

	epID *sql.Stmt

	escalate *sql.Stmt
	epState  *sql.Stmt
}

// A Trigger signals that an alert needs to be processed
type Trigger interface {
	TriggerAlert(int)
}

func NewDB(ctx context.Context, db *sql.DB, logDB alertlog.Store) (*DB, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}

	p := prep.P

	return &DB{
		db:    db,
		logDB: logDB,

		lockSvc:      p(`select 1 from services where id = $1 for update`),
		lockAlertSvc: p(`SELECT 1 FROM services s JOIN alerts a ON a.id = ANY ($1) AND s.id = a.service_id FOR UPDATE`),
		getStatusAndLockSvc: p(`
			SELECT a.status
			FROM services s
			JOIN alerts a on a.id = $1 and a.service_id = s.id
			FOR UPDATE
		`),

		epID: p(`
			SELECT escalation_policy_id
			FROM
				services svc,
				alerts a
			WHERE svc.id = a.service_id
		`),

		insert: p(`
			INSERT INTO alerts (summary, details, service_id, source, status, dedup_key) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at
		`),
		update: p("UPDATE alerts SET status = $2 WHERE id = $1"),
		logs:   p("SELECT timestamp, event, message FROM alert_logs WHERE alert_id = $1"),
		findAllSummary: p(`
			with counts as (
				select count(id), status, service_id
				from alerts
				group by status, service_id
			)
			select distinct
				service_id,
				svc.name,
				(select count from counts c where status = 'triggered' and c.service_id = cn.service_id) triggered,
				(select count from counts c where status = 'active' and c.service_id = cn.service_id) active,
				(select count from counts c where status = 'closed' and c.service_id = cn.service_id) closed
			from counts cn
			join services svc on svc.id = service_id
			order by triggered desc nulls last, active desc nulls last, closed desc nulls last, service_id
			limit 50
		`),

		findMany: p(`
			SELECT
				a.id,
				a.summary,
				a.details,
				a.service_id,
				a.source,
				a.status,
				created_at,
				a.dedup_key
			FROM alerts a
			WHERE a.id = ANY ($1)
		`),
		createUpdNew: p(`
			WITH existing as (
				SELECT id, summary, details, status, source, created_at, false
				FROM alerts
				WHERE service_id = $3 AND dedup_key = $5
			), to_insert as (
				SELECT 1
				EXCEPT
				SELECT 1
				FROM existing
			), inserted as (
				INSERT INTO alerts (
					summary, details, service_id, source, dedup_key
				)
				SELECT $1, $2, $3, $4, $5
				FROM to_insert
				RETURNING id, summary, details, status, source, created_at, true
			)
			SELECT * FROM existing
			UNION
			SELECT * FROM inserted
		`),
		createUpdAck: p(`
			UPDATE alerts a
			SET status = 'active'
			FROM alerts old
			WHERE
				old.id = a.id AND
				a.service_id = $1 AND
				a.dedup_key = $2 AND
				a.status != 'closed'
			RETURNING a.id, a.summary, a.details, old.status, a.created_at
		`),
		createUpdClose: p(`
			UPDATE alerts a
			SET status = 'closed'
			WHERE
				service_id = $1 and
				dedup_key = $2 and
				status != 'closed'
			RETURNING id, summary, details, created_at
		`),

		getCreationTime: p("SELECT created_at FROM alerts WHERE id = $1"),
		getServiceID:    p("SELECT service_id FROM alerts WHERE id = $1"),
		updateByStatusAndService: p(`
			UPDATE
				alerts
			SET
				status = $2
			WHERE
				service_id = $1
			AND (
				$2 > status
			)
		`),
		updateByIDAndStatus: p(`			
			UPDATE alerts
			SET	status = $1
			WHERE
				id = ANY ($2) AND 
				($1 > status)
			RETURNING id
		`),

		escalate: p(`
			UPDATE escalation_policy_state state
			SET force_escalation = true
			WHERE
				state.alert_id = ANY($1) AND
				state.force_escalation = false
			RETURNING state.alert_id
		`),

		epState: p(`
			SELECT alert_id, last_escalation, loop_count, escalation_policy_step_number 
			FROM escalation_policy_state
			WHERE alert_id = ANY ($1)
		`),
	}, prep.Err
}

func (db *DB) EPID(ctx context.Context, alertID int) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return "", err
	}

	row := db.epID.QueryRowContext(ctx, alertID)
	var epID string
	err = row.Scan(&epID)
	if err != nil {
		return "", err
	}
	return epID, nil
}

func (db *DB) canTouchAlert(ctx context.Context, alertID int) error {
	checks := []permission.Checker{
		permission.System,
		permission.Admin,
		permission.User,
	}
	if permission.Service(ctx) {
		var serviceID string
		err := db.getServiceID.QueryRowContext(ctx, alertID).Scan(&serviceID)
		if err != nil {
			return err
		}
		checks = append(checks, permission.MatchService(serviceID))
	}

	return permission.LimitCheckAny(ctx, checks...)
}

func (db *DB) Escalate(ctx context.Context, alertID int, currentLevel int) error {
	_, err := db.EscalateMany(ctx, []int{alertID})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) EscalateMany(ctx context.Context, alertIDs []int) ([]int, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	if len(alertIDs) == 0 {
		return nil, nil
	}

	err = validate.Range("AlertIDs", len(alertIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	ids := sqlutil.IntArray(alertIDs)

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.StmtContext(ctx, db.lockAlertSvc).ExecContext(ctx, ids)
	if err != nil {
		return nil, err
	}

	rows, err := tx.StmtContext(ctx, db.escalate).QueryContext(ctx, ids)
	if err == sql.ErrNoRows {
		log.Debugf(ctx, "escalate alert: no rows matched")
		err = nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedIDs := make([]int, 0, len(alertIDs))
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		updatedIDs = append(updatedIDs, id)
	}

	err = db.logDB.LogManyTx(ctx, tx, updatedIDs, alertlog.TypeEscalationRequest, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return updatedIDs, err
}

func (db *DB) UpdateStatusByService(ctx context.Context, serviceID string, status Status) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return err
	}

	err = validate.OneOf("Status", status, StatusActive, StatusClosed)
	if err != nil {
		return err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	t := alertlog.TypeAcknowledged
	if status == StatusClosed {
		t = alertlog.TypeClosed
	}

	err = db.logDB.LogServiceTx(ctx, tx, serviceID, t, nil)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, db.lockSvc).ExecContext(ctx, serviceID)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, db.updateByStatusAndService).ExecContext(ctx, serviceID, status)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) UpdateManyAlertStatus(ctx context.Context, status Status, alertIDs []int) ([]int, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	if len(alertIDs) == 0 {
		return nil, nil
	}

	err = validate.Many(
		validate.Range("AlertIDs", len(alertIDs), 1, maxBatch),
		validate.OneOf("Status", status, StatusActive, StatusClosed),
	)
	if err != nil {
		return nil, err
	}

	ids := sqlutil.IntArray(alertIDs)

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	t := alertlog.TypeAcknowledged
	if status == StatusClosed {
		t = alertlog.TypeClosed
	}

	var updatedIDs []int

	_, err = tx.StmtContext(ctx, db.lockAlertSvc).ExecContext(ctx, ids)
	if err != nil {
		return nil, err
	}

	rows, err := tx.StmtContext(ctx, db.updateByIDAndStatus).QueryContext(ctx, status, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		updatedIDs = append(updatedIDs, id)
	}

	// Logging Batch Updates for every alertID whose status was updated
	err = db.logDB.LogManyTx(ctx, tx, updatedIDs, t, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return updatedIDs, nil
}

func (db *DB) Create(ctx context.Context, a *Alert) (*Alert, error) {
	n, err := a.Normalize() // validation
	if err != nil {
		return nil, err
	}

	if n.Status == StatusClosed {
		return nil, validation.NewFieldError("Status", "Cannot create a closed alert.")
	}
	err = permission.LimitCheckAny(ctx,
		permission.System,
		permission.Admin,
		permission.User,
		permission.MatchService(a.ServiceID),
	)
	if err != nil {
		return nil, err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.StmtContext(ctx, db.lockSvc).ExecContext(ctx, n.ServiceID)
	if err != nil {
		return nil, err
	}

	n, err = db._create(ctx, tx, *n)
	if err != nil {
		return nil, err
	}

	db.logDB.MustLogTx(ctx, tx, n.ID, alertlog.TypeCreated, nil)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	trace.FromContext(ctx).Annotate(
		[]trace.Attribute{
			trace.StringAttribute("service.id", n.ServiceID),
			trace.Int64Attribute("alert.id", int64(n.ID)),
		},
		"Alert created.",
	)
	ctx = log.WithFields(ctx, log.Fields{"AlertID": n.ID, "ServiceID": n.ServiceID})
	log.Logf(ctx, "Alert created.")

	return n, nil
}
func (db *DB) _create(ctx context.Context, tx *sql.Tx, a Alert) (*Alert, error) {
	row := tx.StmtContext(ctx, db.insert).QueryRowContext(ctx, a.Summary, a.Details, a.ServiceID, a.Source, a.Status, a.DedupKey())
	err := row.Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &a, nil
}
func (db *DB) CreateOrUpdateTx(ctx context.Context, tx *sql.Tx, a *Alert) (*Alert, bool, error) {
	err := permission.LimitCheckAny(ctx,
		permission.System,
		permission.Admin,
		permission.User,
		permission.MatchService(a.ServiceID),
	)
	if err != nil {
		return nil, false, err
	}
	/*
		- if new status is triggered, create or return existing

		- if new status is ack, old is trig, ack and return existing
		- if new status is ack, old is ack, return existing
		- if new status is ack, old is close, return nil

		- if new status is close, old is ack or trig, close, return existing
		- if new status is close, old is close, return nil
	*/

	n, err := a.Normalize() // validation
	if err != nil {
		return nil, false, err
	}

	_, err = tx.StmtContext(ctx, db.lockSvc).ExecContext(ctx, n.ServiceID)
	if err != nil {
		return nil, false, err
	}

	var inserted bool
	var logType alertlog.Type
	switch n.Status {
	case StatusTriggered:
		err = tx.Stmt(db.createUpdNew).
			QueryRowContext(ctx, n.Summary, n.Details, n.ServiceID, n.Source, n.DedupKey()).
			Scan(&n.ID, &n.Summary, &n.Details, &n.Status, &n.Source, &n.CreatedAt, &inserted)
		if !inserted {
			logType = alertlog.TypeDuplicateSupressed
		} else {
			logType = alertlog.TypeCreated
		}
	case StatusActive:
		var oldStatus Status
		err = tx.Stmt(db.createUpdAck).
			QueryRowContext(ctx, n.ServiceID, n.DedupKey()).
			Scan(&n.ID, &n.Summary, &n.Details, &n.CreatedAt, &oldStatus)
		if oldStatus != n.Status {
			logType = alertlog.TypeAcknowledged
		}
	case StatusClosed:
		err = tx.Stmt(db.createUpdClose).
			QueryRowContext(ctx, n.ServiceID, n.DedupKey()).
			Scan(&n.ID, &n.Summary, &n.Details, &n.CreatedAt)
		logType = alertlog.TypeClosed
	}
	if err == sql.ErrNoRows {
		// already closed/doesn't exist
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if logType != "" {
		db.logDB.MustLogTx(ctx, tx, n.ID, logType, nil)
	}

	return n, inserted, nil
}

func (db *DB) CreateOrUpdate(ctx context.Context, a *Alert) (*Alert, error) {
	err := permission.LimitCheckAny(ctx,
		permission.System,
		permission.Admin,
		permission.User,
		permission.MatchService(a.ServiceID),
	)
	if err != nil {
		return nil, err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	n, isNew, err := db.CreateOrUpdateTx(ctx, tx, a)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	if isNew {
		trace.FromContext(ctx).Annotate(
			[]trace.Attribute{
				trace.StringAttribute("service.id", n.ServiceID),
				trace.Int64Attribute("alert.id", int64(n.ID)),
			},
			"Alert created.",
		)
		ctx = log.WithFields(ctx, log.Fields{"AlertID": n.ID, "ServiceID": n.ServiceID})
		log.Logf(ctx, "Alert created.")
	}

	return n, nil
}

func (db *DB) UpdateStatusTx(ctx context.Context, tx *sql.Tx, id int, s Status) error {
	var stat Status
	err := tx.Stmt(db.getStatusAndLockSvc).QueryRowContext(ctx, id).Scan(&stat)
	if err != nil {
		return err
	}
	if stat == StatusClosed {
		return logError{isAlreadyClosed: true, alertID: id, _type: alertlog.TypeClosed, logDB: db.logDB}
	}
	if stat == StatusActive && s == StatusActive {
		return logError{isAlreadyAcknowledged: true, alertID: id, _type: alertlog.TypeAcknowledged, logDB: db.logDB}
	}

	_, err = tx.Stmt(db.update).ExecContext(ctx, id, s)
	if err != nil {
		return err
	}

	if s == StatusClosed {
		db.logDB.MustLogTx(ctx, tx, id, alertlog.TypeClosed, nil)
	} else if s == StatusActive {
		db.logDB.MustLogTx(ctx, tx, id, alertlog.TypeAcknowledged, nil)
	} else if s != StatusTriggered {
		log.Log(ctx, errors.Errorf("unknown/unhandled alert status update: %s", s))
	}

	return nil
}

func (db *DB) UpdateStatus(ctx context.Context, id int, s Status) error {
	err := validate.OneOf("Status", s, StatusTriggered, StatusActive, StatusClosed)
	if err != nil {
		return err
	}
	err = db.canTouchAlert(ctx, id)
	if err != nil {
		return err
	}
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = db.UpdateStatusTx(ctx, tx, id, s)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) GetCreationTime(ctx context.Context, id int) (t time.Time, err error) {
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return t, err
	}

	row := db.getCreationTime.QueryRowContext(ctx, id)
	err = row.Scan(&t)
	return t, err
}

func (db *DB) FindAllSummary(ctx context.Context) ([]Summary, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAllSummary.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var s Summary

	var result []Summary
	var unack, ack, clos sql.NullInt64
	for rows.Next() {
		err = rows.Scan(
			&s.ServiceID,
			&s.ServiceName,
			&unack, &ack, &clos,
		)
		if err != nil {
			return nil, err
		}
		s.Totals.Unack = int(unack.Int64)
		s.Totals.Ack = int(ack.Int64)
		s.Totals.Closed = int(clos.Int64)

		result = append(result, s)
	}

	return result, nil
}

func (db *DB) FindOne(ctx context.Context, id int) (*Alert, error) {
	alerts, err := db.FindMany(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	// If alert is not found
	if len(alerts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &alerts[0], nil
}

func (db *DB) FindMany(ctx context.Context, alertIDs []int) ([]Alert, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}
	if len(alertIDs) == 0 {
		return nil, nil
	}

	err = validate.Range("AlertIDs", len(alertIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	rows, err := db.findMany.QueryContext(ctx, sqlutil.IntArray(alertIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]Alert, 0, len(alertIDs))

	for rows.Next() {
		var a Alert
		err = a.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}

func (db *DB) State(ctx context.Context, alertIDs []int) ([]State, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.Range("AlertIDs", len(alertIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	var t sqlutil.NullTime
	rows, err := db.epState.QueryContext(ctx, sqlutil.IntArray(alertIDs))
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]State, 0, len(alertIDs))
	for rows.Next() {
		var s State
		err = rows.Scan(&s.AlertID, &t, &s.RepeatCount, &s.StepNumber)
		if t.Valid {
			s.LastEscalation = t.Time
		}
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}

	return list, nil
}
