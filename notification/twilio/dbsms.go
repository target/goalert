package twilio

import (
	"context"
	"database/sql"

	"github.com/target/goalert/util"
)

type dbSMS struct {
	db *sql.DB

	lock         *sql.Stmt
	insert       *sql.Stmt
	lookupByCode *sql.Stmt
	lookupLatest *sql.Stmt
	existingCode *sql.Stmt

	lookupByAlert   *sql.Stmt
	lookupSvcByCode *sql.Stmt

	getInUse *sql.Stmt
}

func newDB(ctx context.Context, db *sql.DB) (*dbSMS, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}
	p := prep.P

	//  will register these sql statements by Prepared statements
	return &dbSMS{
		db: db,

		lock: p(`LOCK twilio_sms_callbacks IN SHARE UPDATE EXCLUSIVE MODE`),

		getInUse: p(`
			SELECT cb.code
			FROM twilio_sms_callbacks cb
			WHERE
				phone_number = $1 AND (
					service_id NOTNULL OR
					(SELECT true FROM alerts a WHERE a.id = cb.alert_id AND a.status != 'closed')
				)
		`),

		existingCode: p(`
			SELECT cb.code
			FROM twilio_sms_callbacks cb
			WHERE
				phone_number = $1 AND (
					service_id = $3 OR (
						cb.alert_id = $2 AND
						(SELECT true FROM alerts a WHERE a.id = $2 AND a.status != 'closed')
					)
				)
		`),

		insert: p(`
			INSERT INTO twilio_sms_callbacks (phone_number, callback_id, code, alert_id, service_id)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (phone_number, code) DO UPDATE
			SET
				callback_id = $2,
				alert_id = $4,
				sent_at = now(),
				service_id = $5
		`),

		lookupSvcByCode: p(`
			SELECT callback_id, NULL, name
			FROM twilio_sms_callbacks
			JOIN services svc ON svc.id = service_id
			WHERE phone_number = $1 AND code = $2
		`),
		lookupByCode:  p(`SELECT callback_id, alert_id, NULL FROM twilio_sms_callbacks WHERE phone_number = $1 AND code = $2`),
		lookupByAlert: p(`SELECT callback_id, alert_id, NULL FROM twilio_sms_callbacks WHERE phone_number = $1 AND alert_id = $2`),

		lookupLatest: p(`
			SELECT callback_id, alert_id, NULL
			FROM twilio_sms_callbacks
			WHERE phone_number = $1 AND alert_id NOTNULL
			ORDER BY sent_at DESC
			LIMIT 1
		`),
	}, prep.Err
}

func (db *dbSMS) insertDB(ctx context.Context, phoneNumber, callbackID string, alertID int, serviceID string) (int, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	_, err = tx.StmtContext(ctx, db.lock).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	aID := sql.NullInt64{Int64: int64(alertID)}
	sID := sql.NullString{String: serviceID}
	if alertID != 0 {
		aID.Valid = true
	}
	if serviceID != "" {
		sID.Valid = true
	}

	var existingCode sql.NullInt64
	err = tx.StmtContext(ctx, db.existingCode).QueryRowContext(ctx, phoneNumber, aID, sID).Scan(&existingCode)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return 0, err
	}
	if existingCode.Valid {
		return int(existingCode.Int64), nil
	}

	rows, err := tx.StmtContext(ctx, db.getInUse).QueryContext(ctx, phoneNumber)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	m := make(map[int]struct{})
	for rows.Next() {
		var code int
		err = rows.Scan(&code)
		if err != nil {
			return 0, err
		}
		m[code] = struct{}{}
	}

	code := 1
	if serviceID != "" {
		code = 100
	}
	for {
		if _, ok := m[code]; !ok {
			break
		}
		code++
	}

	_, err = tx.StmtContext(ctx, db.insert).ExecContext(ctx, phoneNumber, callbackID, code, aID, sID)
	if err != nil {
		return 0, err
	}

	return code, tx.Commit()
}

type codeInfo struct {
	ServiceName string
	AlertID     int
	CallbackID  string
}

func (c *codeInfo) scanFrom(row *sql.Row) error {
	var aID sql.NullInt64
	var svcName sql.NullString
	err := row.Scan(&c.CallbackID, &aID, &svcName)
	if err != nil {
		return err
	}
	c.ServiceName = svcName.String
	c.AlertID = int(aID.Int64)

	return nil
}

func (db *dbSMS) LookupByCode(ctx context.Context, phoneNumber string, code int) (*codeInfo, error) {
	var row *sql.Row
	if code != 0 {
		row = db.lookupByCode.QueryRowContext(ctx, phoneNumber, code)
	} else {
		row = db.lookupLatest.QueryRowContext(ctx, phoneNumber)
	}

	info := &codeInfo{}
	err := info.scanFrom(row)
	return info, err
}
func (db *dbSMS) LookupByAlertID(ctx context.Context, phoneNumber string, searchID int) (*codeInfo, error) {
	row := db.lookupByAlert.QueryRowContext(ctx, phoneNumber, searchID)

	info := &codeInfo{}
	err := info.scanFrom(row)
	return info, err
}
func (db *dbSMS) LookupSvcByCode(ctx context.Context, phoneNumber string, code int) (*codeInfo, error) {
	row := db.lookupSvcByCode.QueryRowContext(ctx, phoneNumber, code)

	info := &codeInfo{}
	err := info.scanFrom(row)
	return info, err
}
