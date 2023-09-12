package alert

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

const maxBatch = 500

type Store struct {
	db    *sql.DB
	logDB *alertlog.Store

	insert       *sql.Stmt
	update       *sql.Stmt
	logs         *sql.Stmt
	findMany     *sql.Stmt
	getServiceID *sql.Stmt

	lockSvc      *sql.Stmt
	lockAlertSvc *sql.Stmt

	getStatusAndLockSvc *sql.Stmt

	createUpdNew   *sql.Stmt
	createUpdAck   *sql.Stmt
	createUpdClose *sql.Stmt

	updateByStatusAndService *sql.Stmt
	updateByIDAndStatus      *sql.Stmt

	noStepsBySvc *sql.Stmt

	epID *sql.Stmt

	escalate *sql.Stmt
	epState  *sql.Stmt
	svcInfo  *sql.Stmt
}

// A Trigger signals that an alert needs to be processed
type Trigger interface {
	TriggerAlert(int)
}

func NewStore(ctx context.Context, db *sql.DB, logDB *alertlog.Store) (*Store, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}

	p := prep.P

	return &Store{
		db:    db,
		logDB: logDB,

		noStepsBySvc: p(`
			SELECT coalesce(
				(SELECT true
				FROM escalation_policies pol
				JOIN services svc ON svc.id = $1
				WHERE
					pol.id = svc.escalation_policy_id
					AND pol.step_count = 0)
			, false)
		`),

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

		getServiceID: p("SELECT service_id FROM alerts WHERE id = $1"),
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
			FROM alerts as a, services as svc
			WHERE
				state.alert_id = ANY($1) AND
				state.force_escalation = false AND
				a.id = state.alert_id AND
				svc.id = a.service_id AND
				svc.maintenance_expires_at ISNULL
			RETURNING state.alert_id
		`),

		epState: p(`
			SELECT alert_id, last_escalation, loop_count, escalation_policy_step_number 
			FROM escalation_policy_state
			WHERE alert_id = ANY ($1)
		`),

		svcInfo: p(`
			SELECT
				name,
				(SELECT count(*) FROM alerts WHERE service_id = $1 AND status = 'triggered')
			FROM services
			WHERE id = $1
		`),
	}, prep.Err
}

// ServiceInfo will return the name of the given service ID as well as the current number
// of unacknowledged alerts.
func (s *Store) ServiceInfo(ctx context.Context, serviceID string) (string, int, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return "", 0, err
	}

	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return "", 0, err
	}

	var name string
	var count int
	err = s.svcInfo.QueryRowContext(ctx, serviceID).Scan(&name, &count)
	if err != nil {
		return "", 0, err
	}

	return name, count, nil
}

func (s *Store) EPID(ctx context.Context, alertID int) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return "", err
	}

	row := s.epID.QueryRowContext(ctx, alertID)
	var epID string
	err = row.Scan(&epID)
	if err != nil {
		return "", err
	}
	return epID, nil
}

func (s *Store) canTouchAlert(ctx context.Context, alertID int) error {
	checks := []permission.Checker{
		permission.System,
		permission.Admin,
		permission.User,
	}
	if permission.Service(ctx) {
		var serviceID string
		err := s.getServiceID.QueryRowContext(ctx, alertID).Scan(&serviceID)
		if err != nil {
			return err
		}
		checks = append(checks, permission.MatchService(serviceID))
	}

	return permission.LimitCheckAny(ctx, checks...)
}

// EscalateAsOf will request escalation for the given alert ID as-of the given time.
//
// An error will be returned if the alert is already closed, if the service is
// in maintenance mode, there are no steps on the escalation policy, or if the
// alert has already been escalated since the given time.
func (s *Store) EscalateAsOf(ctx context.Context, id int, t time.Time) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "escalate alert", tx)

	lck, err := gadb.New(tx).LockOneAlertService(ctx, int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return validation.NewGenericError("alert not found")
	}
	if err != nil {
		return fmt.Errorf("lock alert: %w", err)
	}
	if lck.Status == gadb.EnumAlertStatusClosed {
		return logError{isAlreadyClosed: true, alertID: id, _type: alertlog.TypeClosed, logDB: s.logDB}
	}
	if lck.IsMaintMode {
		return validation.NewGenericError("service is in maintenance mode")
	}

	if t.IsZero() {
		t, err = gadb.New(tx).Now(ctx)
		if err != nil {
			return fmt.Errorf("get current time: %w", err)
		}
	}

	ok, err := gadb.New(tx).RequestAlertEscalationByTime(ctx, gadb.RequestAlertEscalationByTimeParams{
		AlertID: int64(id),
		Column2: t,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("request escalation: %w", err)
	}

	if !ok {
		hasEP, err := gadb.New(tx).AlertHasEPState(ctx, int64(id))
		if err != nil {
			return fmt.Errorf("check ep state: %w", err)
		}

		if !hasEP {
			return validation.NewGenericError("alert escalation policy is empty")
		}

		return validation.NewGenericError("alert has already escalated")
	}

	err = s.logDB.LogTx(ctx, tx, id, alertlog.TypeEscalationRequest, nil)
	if err != nil {
		return fmt.Errorf("log escalation request: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Store) Escalate(ctx context.Context, alertID int, currentLevel int) error {
	return s.EscalateAsOf(ctx, alertID, time.Time{})
}

func (s *Store) EscalateMany(ctx context.Context, alertIDs []int) ([]int, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	if len(alertIDs) == 0 {
		return nil, nil
	}

	err = validate.Range("AlertIDs", len(alertIDs), 1, 1)
	if err != nil {
		return nil, err
	}
	err = s.EscalateAsOf(ctx, alertIDs[0], time.Time{})
	if err != nil {
		return nil, err
	}

	return alertIDs, nil
}

func (s *Store) UpdateStatusByService(ctx context.Context, serviceID string, status Status) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "alert: update status by service", tx)

	t := alertlog.TypeAcknowledged
	if status == StatusClosed {
		t = alertlog.TypeClosed
	}

	err = s.logDB.LogServiceTx(ctx, tx, serviceID, t, nil)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, s.lockSvc).ExecContext(ctx, serviceID)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, s.updateByStatusAndService).ExecContext(ctx, serviceID, status)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) UpdateManyAlertStatus(ctx context.Context, status Status, alertIDs []int, logMeta interface{}) ([]int, error) {
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer sqlutil.Rollback(ctx, "alert: update status", tx)

	t := alertlog.TypeAcknowledged
	if status == StatusClosed {
		t = alertlog.TypeClosed
	}

	var updatedIDs []int

	_, err = tx.StmtContext(ctx, s.lockAlertSvc).ExecContext(ctx, ids)
	if err != nil {
		return nil, err
	}

	rows, err := tx.StmtContext(ctx, s.updateByIDAndStatus).QueryContext(ctx, status, ids)
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
	err = s.logDB.LogManyTx(ctx, tx, updatedIDs, t, logMeta)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return updatedIDs, nil
}

func (s *Store) Create(ctx context.Context, a *Alert) (*Alert, error) {
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer sqlutil.Rollback(ctx, "alert: create", tx)

	_, err = tx.StmtContext(ctx, s.lockSvc).ExecContext(ctx, n.ServiceID)
	if err != nil {
		return nil, err
	}

	n, meta, err := s._create(ctx, tx, *n)
	if err != nil {
		return nil, err
	}

	s.logDB.MustLogTx(ctx, tx, n.ID, alertlog.TypeCreated, meta)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	ctx = log.WithFields(ctx, log.Fields{"AlertID": n.ID, "ServiceID": n.ServiceID})
	log.Logf(ctx, "Alert created.")
	metricCreatedTotal.Inc()

	return n, nil
}

func (s *Store) _create(ctx context.Context, tx *sql.Tx, a Alert) (*Alert, *alertlog.CreatedMetaData, error) {
	var meta alertlog.CreatedMetaData
	row := tx.StmtContext(ctx, s.insert).QueryRowContext(ctx, a.Summary, a.Details, a.ServiceID, a.Source, a.Status, a.DedupKey())
	err := row.Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	err = tx.StmtContext(ctx, s.noStepsBySvc).QueryRowContext(ctx, a.ServiceID).Scan(&meta.EPNoSteps)
	if err != nil {
		return nil, nil, err
	}

	return &a, &meta, nil
}

// CreateOrUpdateTx returns `isNew` to indicate if the returned alert was a new one.
// It is the caller's responsibility to log alert creation if the transaction is committed (and isNew is true).
func (s *Store) CreateOrUpdateTx(ctx context.Context, tx *sql.Tx, a *Alert) (*Alert, bool, error) {
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

	_, err = tx.StmtContext(ctx, s.lockSvc).ExecContext(ctx, n.ServiceID)
	if err != nil {
		return nil, false, err
	}

	var inserted bool
	var logType alertlog.Type
	var meta interface{}
	switch n.Status {
	case StatusTriggered:
		var m alertlog.CreatedMetaData
		err = tx.Stmt(s.createUpdNew).
			QueryRowContext(ctx, n.Summary, n.Details, n.ServiceID, n.Source, n.DedupKey()).
			Scan(&n.ID, &n.Summary, &n.Details, &n.Status, &n.Source, &n.CreatedAt, &inserted)
		if !inserted {
			logType = alertlog.TypeDuplicateSupressed
		} else {
			logType = alertlog.TypeCreated
			stepErr := tx.StmtContext(ctx, s.noStepsBySvc).QueryRowContext(ctx, n.ServiceID).Scan(&m.EPNoSteps)
			if stepErr != nil {
				return nil, false, err
			}
		}
		meta = &m
	case StatusActive:
		var oldStatus Status
		err = tx.Stmt(s.createUpdAck).
			QueryRowContext(ctx, n.ServiceID, n.DedupKey()).
			Scan(&n.ID, &n.Summary, &n.Details, &oldStatus, &n.CreatedAt)
		if oldStatus != n.Status {
			logType = alertlog.TypeAcknowledged
		}
	case StatusClosed:
		err = tx.Stmt(s.createUpdClose).
			QueryRowContext(ctx, n.ServiceID, n.DedupKey()).
			Scan(&n.ID, &n.Summary, &n.Details, &n.CreatedAt)
		logType = alertlog.TypeClosed
	}
	if errors.Is(err, sql.ErrNoRows) {
		// already closed/doesn't exist
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if logType != "" {
		s.logDB.MustLogTx(ctx, tx, n.ID, logType, meta)
	}

	return n, inserted, nil
}

// CreateOrUpdate will create an alert or log a "duplicate suppressed message" if
// Status is Triggered. If Status is Closed, it will close and return the result.
//
// In the case that Status is closed but a matching alert is not present, nil is returned.
// Otherwise the current alert is returned.
func (s *Store) CreateOrUpdate(ctx context.Context, a *Alert) (*Alert, bool, error) {
	err := permission.LimitCheckAny(ctx,
		permission.System,
		permission.Admin,
		permission.User,
		permission.MatchService(a.ServiceID),
	)
	if err != nil {
		return nil, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer sqlutil.Rollback(ctx, "alert: upsert", tx)

	n, isNew, err := s.CreateOrUpdateTx(ctx, tx, a)
	if err != nil {
		return nil, false, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, false, err
	}
	if n == nil {
		return nil, false, nil
	}
	if isNew {
		ctx = log.WithFields(ctx, log.Fields{"AlertID": n.ID, "ServiceID": n.ServiceID})
		log.Logf(ctx, "Alert created.")
		metricCreatedTotal.Inc()
	}

	return n, isNew, nil
}

func (s *Store) UpdateStatusTx(ctx context.Context, tx *sql.Tx, id int, stat Status) error {
	var _stat Status
	err := tx.Stmt(s.getStatusAndLockSvc).QueryRowContext(ctx, id).Scan(&_stat)
	if err != nil {
		return err
	}
	if _stat == StatusClosed {
		return logError{isAlreadyClosed: true, alertID: id, _type: alertlog.TypeClosed, logDB: s.logDB}
	}
	if _stat == StatusActive && stat == StatusActive {
		return logError{isAlreadyAcknowledged: true, alertID: id, _type: alertlog.TypeAcknowledged, logDB: s.logDB}
	}

	_, err = tx.Stmt(s.update).ExecContext(ctx, id, stat)
	if err != nil {
		return err
	}

	if stat == StatusClosed {
		s.logDB.MustLogTx(ctx, tx, id, alertlog.TypeClosed, nil)
	} else if stat == StatusActive {
		s.logDB.MustLogTx(ctx, tx, id, alertlog.TypeAcknowledged, nil)
	} else if stat != StatusTriggered {
		log.Log(ctx, errors.Errorf("unknown/unhandled alert status update: %s", stat))
	}

	return nil
}

func (s *Store) UpdateStatus(ctx context.Context, id int, stat Status) error {
	err := validate.OneOf("Status", stat, StatusTriggered, StatusActive, StatusClosed)
	if err != nil {
		return err
	}
	err = s.canTouchAlert(ctx, id)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "alert: update status", tx)

	err = s.UpdateStatusTx(ctx, tx, id, stat)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) FindOne(ctx context.Context, id int) (*Alert, error) {
	alerts, err := s.FindMany(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	// If alert is not found
	if len(alerts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &alerts[0], nil
}

func (s *Store) FindMany(ctx context.Context, alertIDs []int) ([]Alert, error) {
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

	rows, err := s.findMany.QueryContext(ctx, sqlutil.IntArray(alertIDs))
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

func (s *Store) State(ctx context.Context, alertIDs []int) ([]State, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.Range("AlertIDs", len(alertIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	var t sqlutil.NullTime
	rows, err := s.epState.QueryContext(ctx, sqlutil.IntArray(alertIDs))
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]State, 0, len(alertIDs))
	for rows.Next() {
		var s State
		err = rows.Scan(&s.ID, &t, &s.RepeatCount, &s.StepNumber)
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

func (s *Store) Feedback(ctx context.Context, alertIDs []int) ([]Feedback, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.Range("AlertIDs", len(alertIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	ids := make([]int32, len(alertIDs))
	for _, id := range alertIDs {
		ids = append(ids, int32(id))
	}

	rows, err := gadb.New(s.db).AlertFeedback(ctx, ids)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []Feedback

	for _, r := range rows {
		result = append(result, Feedback{ID: int(r.AlertID), NoiseReason: r.NoiseReason})
	}
	return result, nil
}

func (s Store) UpdateManyAlertFeedback(ctx context.Context, noiseReason string, alertIDs []int) ([]int, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.Text("NoiseReason", noiseReason, 1, 255)
	if err != nil {
		return nil, err
	}

	// GraphQL generates type of int[], while sqlc
	// expects an int64[] as a result of the unnest function
	ids := make([]int64, len(alertIDs))
	for i, v := range alertIDs {
		ids[i] = int64(v)
	}

	res, err := gadb.New(s.db).SetManyAlertFeedback(ctx, gadb.SetManyAlertFeedbackParams{
		AlertIds:    ids,
		NoiseReason: noiseReason,
	})
	if err != nil {
		return nil, err
	}

	// cast back to []int
	updatedIDs := make([]int, len(res))
	for i, v := range res {
		updatedIDs[i] = int(v)
	}

	return updatedIDs, nil
}

func (s Store) UpdateFeedback(ctx context.Context, feedback *Feedback) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return err
	}

	err = validate.Text("NoiseReason", feedback.NoiseReason, 1, 255)
	if err != nil {
		return err
	}

	err = gadb.New(s.db).SetAlertFeedback(ctx, gadb.SetAlertFeedbackParams{
		AlertID:     int64(feedback.ID),
		NoiseReason: feedback.NoiseReason,
	})
	if err != nil {
		return err
	}

	return nil
}
