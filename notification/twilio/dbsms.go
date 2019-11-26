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

	lookupByAlert *sql.Stmt

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
			JOIN alerts a ON a.id = cb.alert_id AND a.status != 'closed'
			WHERE phone_number = $1
		`),

		existingCode: p(`
			SELECT cb.code
			FROM twilio_sms_callbacks cb
			JOIN alerts a ON a.id = cb.alert_id AND a.status != 'closed'
			WHERE phone_number = $1 AND cb.alert_id = $2
		`),

		insert: p(`
			INSERT INTO twilio_sms_callbacks (phone_number, callback_id, code, alert_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (phone_number, code) DO UPDATE
			SET
				callback_id = $2,
				alert_id = $4,
				sent_at = now()
		`),

		lookupByCode:  p(`SELECT callback_id, alert_id FROM twilio_sms_callbacks WHERE phone_number = $1 AND code = $2`),
		lookupByAlert: p(`SELECT callback_id FROM twilio_sms_callbacks WHERE phone_number = $1 AND alert_id = $2`),

		lookupLatest: p(`
			SELECT callback_id, alert_id
			FROM twilio_sms_callbacks
			WHERE phone_number = $1
			ORDER BY sent_at DESC
			LIMIT 1
		`),
	}, prep.Err
}

func (db *dbSMS) insertDB(ctx context.Context, phoneNumber string, callbackID string, alertID int) (int, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	_, err = tx.StmtContext(ctx, db.lock).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	var existingCode sql.NullInt64
	err = tx.StmtContext(ctx, db.existingCode).QueryRowContext(ctx, phoneNumber, alertID).Scan(&existingCode)
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
	for {
		if _, ok := m[code]; !ok {
			break
		}
		code++
	}

	_, err = tx.StmtContext(ctx, db.insert).ExecContext(ctx, phoneNumber, callbackID, code, alertID)
	if err != nil {
		return 0, err
	}

	return code, tx.Commit()
}

func (db *dbSMS) LookupByCode(ctx context.Context, phoneNumber string, code int) (callbackID string, alertID int, err error) {
	var row *sql.Row
	if code != 0 {
		row = db.lookupByCode.QueryRowContext(ctx, phoneNumber, code)
	} else {
		row = db.lookupLatest.QueryRowContext(ctx, phoneNumber)
	}
	err = row.Scan(&callbackID, &alertID)
	return callbackID, alertID, err
}
func (db *dbSMS) LookupByAlertID(ctx context.Context, phoneNumber string, searchID int) (callbackID string, alertID int, err error) {
	err = db.lookupByAlert.QueryRowContext(ctx, phoneNumber, searchID).Scan(&callbackID)
	return callbackID, searchID, err
}
