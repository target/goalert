package notification

import (
	"context"
	cRand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const minTimeBetweenTests = time.Minute

type Store interface {
	SendContactMethodTest(ctx context.Context, cmID string) error
	SendContactMethodVerification(ctx context.Context, cmID string, resend bool) error
	VerifyContactMethod(ctx context.Context, cmID string, code int) ([]string, error)
	CodeExpiration(ctx context.Context, cmID string) (*time.Time, error)
	Code(ctx context.Context, id string) (int, error)
}

var _ Store = &DB{}

type DB struct {
	db                     *sql.DB
	getCMUserID            *sql.Stmt
	setVerificationCode    *sql.Stmt
	verifyVerificationCode *sql.Stmt
	enableContactMethods   *sql.Stmt
	insertTestNotification *sql.Stmt
	updateLastSendTime     *sql.Stmt
	codeExpiration         *sql.Stmt
	getCode                *sql.Stmt
	sendTestLock           *sql.Stmt

	rand *rand.Rand
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	var seed int64
	err := binary.Read(cRand.Reader, binary.BigEndian, &seed)
	if err != nil {
		return nil, errors.Wrap(err, "generate random seed")
	}

	return &DB{
		db: db,

		rand: rand.New(rand.NewSource(seed)),

		getCMUserID: p.P(`select user_id from user_contact_methods where id = $1`),

		sendTestLock: p.P(`lock outgoing_messages, user_contact_methods in row exclusive mode`),

		getCode: p.P(`
			select code
			from user_verification_codes
			where id = $1
		`),

		codeExpiration: p.P(`
			select expires_at
			from user_verification_codes v
			join user_contact_methods cm on cm.id = $1
			where v.user_id = cm.user_id and v.contact_method_value = cm.value
		`),

		setVerificationCode: p.P(`
			insert into user_verification_codes (id, user_id, contact_method_value, code, expires_at, send_to)
			select
				$1,
				cm.user_id,
				cm.value,
				$3,
				now() + cast($4 as interval),
				$2
			from user_contact_methods cm
			where id = $2
			on conflict (user_id, contact_method_value) do update
			set
				send_to = $2,
				expires_at = case when $5 then user_verification_codes.expires_at else EXCLUDED.expires_at end,
				code = case when $5 then user_verification_codes.code else EXCLUDED.code end
		`),
		verifyVerificationCode: p.P(`
			delete from user_verification_codes v
			using user_contact_methods cm
			where
				cm.id = $1 and
				v.contact_method_value = cm.value and
				v.user_id = cm.user_id and
				v.code = $2
			returning cm.value
		`),

		enableContactMethods: p.P(`
			update user_contact_methods
			set disabled = false
			where user_id = $1 and value = $2
			returning id
		`),

		updateLastSendTime: p.P(`
			update user_contact_methods
			set last_test_verify_at = now()
			where
				id = $1 and
				(
					last_test_verify_at + cast($2 as interval) < now()
					or
					last_test_verify_at isnull
				)
		`),

		insertTestNotification: p.P(`
			insert into outgoing_messages (id, message_type, contact_method_id, user_id)
			select
				$1,
				'test_notification',
				$2,
				cm.user_id
			from user_contact_methods cm
			where cm.id = $2
		`),
	}, p.Err
}

func (db *DB) cmUserID(ctx context.Context, id string) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return "", err
	}

	err = validate.UUID("ContactMethodID", id)
	if err != nil {
		return "", err
	}

	var userID string
	err = db.getCMUserID.QueryRowContext(ctx, id).Scan(&userID)
	if err != nil {
		return "", err
	}

	// only admin or same-user can verify
	err = permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (db *DB) Code(ctx context.Context, id string) (int, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return 0, err
	}
	err = validate.UUID("VerificationCodeID", id)
	if err != nil {
		return 0, err
	}

	var code int
	err = db.getCode.QueryRowContext(ctx, id).Scan(&code)
	return code, err
}

func (db *DB) CodeExpiration(ctx context.Context, id string) (t *time.Time, err error) {
	_, err = db.cmUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = db.codeExpiration.QueryRowContext(ctx, id).Scan(&t)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (db *DB) SendContactMethodTest(ctx context.Context, id string) error {
	_, err := db.cmUserID(ctx, id)
	if err != nil {
		return err
	}
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock outgoing_messages first, before we modify user_contact methods
	// to prevent deadlock.
	_, err = tx.Stmt(db.sendTestLock).ExecContext(ctx)
	if err != nil {
		return err
	}

	r, err := tx.Stmt(db.updateLastSendTime).ExecContext(ctx, id, fmt.Sprintf("%f seconds", minTimeBetweenTests.Seconds()))
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return validation.NewFieldError("ContactMethod", "test message rate-limit exceeded")
	}

	vID := uuid.NewV4().String()
	_, err = tx.Stmt(db.insertTestNotification).ExecContext(ctx, vID, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) SendContactMethodVerification(ctx context.Context, id string, resend bool) error {
	_, err := db.cmUserID(ctx, id)
	if err != nil {
		return err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	r, err := tx.Stmt(db.updateLastSendTime).ExecContext(ctx, id, fmt.Sprintf("%f seconds", minTimeBetweenTests.Seconds()))
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return validation.NewFieldError("ContactMethod", "test message rate-limit exceeded")
	}

	vID := uuid.NewV4().String()
	code := db.rand.Intn(900000) + 100000
	_, err = tx.Stmt(db.setVerificationCode).ExecContext(ctx, vID, id, code, fmt.Sprintf("%f seconds", (15*time.Minute).Seconds()), resend)
	if err != nil {
		return errors.Wrap(err, "set verification code")
	}

	return tx.Commit()
}

func (db *DB) VerifyContactMethod(ctx context.Context, cmID string, code int) ([]string, error) {
	userID, err := db.cmUserID(ctx, cmID)
	if err != nil {
		return nil, err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var cmValue string
	err = db.verifyVerificationCode.QueryRowContext(ctx, cmID, code).Scan(&cmValue)
	if err == sql.ErrNoRows {
		return nil, validation.NewFieldError("Code", "unrecognized code")
	}
	if err != nil {
		return nil, err
	}

	rows, err := db.enableContactMethods.QueryContext(ctx, userID, cmValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, tx.Commit()
}
