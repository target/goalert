package notification

import (
	"context"
	cRand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const minTimeBetweenTests = time.Minute

type Store interface {
	SendContactMethodTest(ctx context.Context, cmID string) error
	SendContactMethodVerification(ctx context.Context, cmID string) error
	VerifyContactMethod(ctx context.Context, cmID string, code int) error
	Code(ctx context.Context, id string) (int, error)
	FindManyMessageStatuses(ctx context.Context, ids ...string) ([]MessageStatus, error)
}

var _ Store = &DB{}

type DB struct {
	db                           *sql.DB
	getCMUserID                  *sql.Stmt
	setVerificationCode          *sql.Stmt
	verifyAndEnableContactMethod *sql.Stmt
	insertTestNotification       *sql.Stmt
	updateLastSendTime           *sql.Stmt
	getCode                      *sql.Stmt
	isDisabled                   *sql.Stmt
	sendTestLock                 *sql.Stmt
	findManyMessageStatuses      *sql.Stmt

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

		isDisabled: p.P(`
			select disabled
			from user_contact_methods
			where id = $1
		`),

		// should result in sending a verification code to the specified contact method
		setVerificationCode: p.P(`
			insert into user_verification_codes (id, contact_method_id, code, expires_at)
			values ($1, $2, $3, NOW() + '15 minutes'::interval)
			on conflict (contact_method_id) do update
			set
				sent = false,
				expires_at = EXCLUDED.expires_at
		`),

		// should reactivate a contact method if specified code matches what was set
		verifyAndEnableContactMethod: p.P(`
			with v as (
				delete from user_verification_codes
				where contact_method_id = $1 and code = $2
				returning contact_method_id id
			)
			update user_contact_methods cm
			set disabled = false
			from v
			where cm.id = v.id
			returning cm.id
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

		findManyMessageStatuses: p.P(`
				select
					id,
					last_status,
					status_details,
					provider_msg_id,
					provider_seq,
					next_retry_at notnull
				from outgoing_messages
				where id = any($1)
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
	if err == sql.ErrNoRows {
		return "", validation.NewFieldError("ContactMethodID", "does not exist")
	}
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
	_, err = tx.StmtContext(ctx, db.sendTestLock).ExecContext(ctx)
	if err != nil {
		return err
	}

	var isDisabled bool
	err = tx.StmtContext(ctx, db.isDisabled).QueryRowContext(ctx, id).Scan(&isDisabled)
	if err != nil {
		return err
	}
	if isDisabled {
		return validation.NewFieldError("ContactMethod", "contact method disabled")
	}

	r, err := tx.StmtContext(ctx, db.updateLastSendTime).ExecContext(ctx, id, fmt.Sprintf("%f seconds", minTimeBetweenTests.Seconds()))
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
	_, err = tx.StmtContext(ctx, db.insertTestNotification).ExecContext(ctx, vID, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) SendContactMethodVerification(ctx context.Context, cmID string) error {
	_, err := db.cmUserID(ctx, cmID)
	if err != nil {
		return err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	r, err := tx.StmtContext(ctx, db.updateLastSendTime).ExecContext(ctx, cmID, fmt.Sprintf("%f seconds", minTimeBetweenTests.Seconds()))
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return validation.NewFieldError("ContactMethod", fmt.Sprintf("Too many messages! Please try again in %.0f minute(s)", minTimeBetweenTests.Minutes()))
	}

	vcID := uuid.NewV4().String()
	code := db.rand.Intn(900000) + 100000
	_, err = tx.StmtContext(ctx, db.setVerificationCode).ExecContext(ctx, vcID, cmID, code)
	if err != nil {
		return errors.Wrap(err, "set verification code")
	}

	return tx.Commit()
}

func (db *DB) VerifyContactMethod(ctx context.Context, cmID string, code int) error {
	_, err := db.cmUserID(ctx, cmID)
	if err != nil {
		return err
	}

	res, err := db.verifyAndEnableContactMethod.ExecContext(ctx, cmID, code)
	if err == sql.ErrNoRows {
		return validation.NewFieldError("code", "invalid code")
	}
	if err != nil {
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return validation.NewFieldError("code", "invalid code")
	}

	// NOTE: maintain a record of consent/dissent
	logCtx := log.WithFields(ctx, log.Fields{
		"contactMethodID": cmID,
	})

	log.Logf(logCtx, "Contact method ENABLED/VERIFIED.")

	return nil
}

func (db *DB) FindManyMessageStatuses(ctx context.Context, ids ...string) ([]MessageStatus, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	err = validate.ManyUUID("IDs", ids, search.MaxResults)
	if err != nil {
		return nil, err
	}

	rows, err := db.findManyMessageStatuses.QueryContext(ctx, sqlutil.UUIDArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MessageStatus
	var s MessageStatus
	for rows.Next() {
		var lastStatus string
		var hasNextRetry bool
		var providerMsgID sql.NullString
		err = rows.Scan(&s.ID, &lastStatus, &s.Details, &providerMsgID, &s.Sequence, &hasNextRetry)
		if err != nil {
			return nil, err
		}
		s.ProviderMessageID = providerMsgID.String

		switch lastStatus {
		case "pending", "sending", "queued_remotely":
			s.State = MessageStateActive
		case "sent":
			s.State = MessageStateSent
		case "delivered":
			s.State = MessageStateDelivered
		case "failed", "bundled": // bundled message was not sent (replaced) and should never be re-sent
			// temporary if retry
			if hasNextRetry {
				s.State = MessageStateFailedTemp
			} else {
				s.State = MessageStateFailedPerm
			}
		default:
			return nil, fmt.Errorf("unknown last_status %s", lastStatus)
		}
		s.LastStatus = MessageLastStatus(lastStatus)
		result = append(result, s)
	}

	return result, nil
}
