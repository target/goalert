package notification

import (
	"context"
	cRand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const minTimeBetweenTests = time.Minute

type Store struct {
	db                           *sql.DB
	getCMUserID                  *sql.Stmt
	setVerificationCode          *sql.Stmt
	verifyAndEnableContactMethod *sql.Stmt
	insertTestNotification       *sql.Stmt
	updateLastSendTime           *sql.Stmt
	getCode                      *sql.Stmt
	isDisabled                   *sql.Stmt
	sendTestLock                 *sql.Stmt

	rand *rand.Rand
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	var seed int64
	err := binary.Read(cRand.Reader, binary.BigEndian, &seed)
	if err != nil {
		return nil, errors.Wrap(err, "generate random seed")
	}

	return &Store{
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
	}, p.Err
}

// OriginalMessageStatus will return the status of the first alert notification sent to `dest` for the given `alertID`.
func (s *Store) OriginalMessageStatus(ctx context.Context, alertID int, dst Dest) (*SendResult, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return nil, err
	}

	var cmID, chanID uuid.NullUUID
	if dst.ID.IsUserCM() {
		cmID.UUID, cmID.Valid = dst.ID.UUID(), true
	} else {
		chanID.UUID, chanID.Valid = dst.ID.UUID(), true
	}

	row, err := gadb.New(s.db).NfyOriginalMessageStatus(ctx, gadb.NfyOriginalMessageStatusParams{
		AlertID:         sql.NullInt64{Valid: true, Int64: int64(alertID)},
		ContactMethodID: cmID,
		ChannelID:       chanID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return outgoingMessageToSendResult(row.OutgoingMessage, row.CmDest, row.ChDest)
}

func outgoingMessageToSendResult(msg gadb.OutgoingMessage, cm, ch gadb.NullDestV1) (*SendResult, error) {
	res := SendResult{
		ID:                msg.ID.String(),
		ProviderMessageID: msg.ProviderMsgID,
	}

	switch {
	case cm.Valid:
		res.DestType = DestV1TypeToDestType(cm.DestV1.Type)
	case ch.Valid:
		res.DestType = DestV1TypeToDestType(ch.DestV1.Type)
	}

	state, err := messageStateFromStatus(string(msg.LastStatus), msg.NextRetryAt.Valid)
	if err != nil {
		return nil, err
	}

	res.Status = Status{
		State:    state,
		Details:  msg.StatusDetails,
		Sequence: int(msg.ProviderSeq),
		SrcValue: msg.SrcValue.String,
	}
	if msg.LastStatusAt.Valid {
		res.Status.age = msg.LastStatusAt.Time.Sub(msg.CreatedAt)
	}

	return &res, nil
}

func (s *Store) cmUserID(ctx context.Context, id string) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return "", err
	}

	err = validate.UUID("ContactMethodID", id)
	if err != nil {
		return "", err
	}

	var userID string
	err = s.getCMUserID.QueryRowContext(ctx, id).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
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

func (s *Store) Code(ctx context.Context, id string) (int, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return 0, err
	}
	err = validate.UUID("VerificationCodeID", id)
	if err != nil {
		return 0, err
	}

	var code int
	err = s.getCode.QueryRowContext(ctx, id).Scan(&code)
	return code, err
}

func (s *Store) SendContactMethodTest(ctx context.Context, id string) error {
	cmUserID, err := s.cmUserID(ctx, id)
	if err != nil {
		return err
	}

	// due to potential regulations around consent with phone calls and SMS, we
	// only allow users to send test messages to their own contact methods
	err = permission.LimitCheckAny(ctx, permission.MatchUser(cmUserID))
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "notification: send test message", tx)

	// Lock outgoing_messages first, before we modify user_contact methods
	// to prevent deadlock.
	_, err = tx.StmtContext(ctx, s.sendTestLock).ExecContext(ctx)
	if err != nil {
		return err
	}

	var isDisabled bool
	err = tx.StmtContext(ctx, s.isDisabled).QueryRowContext(ctx, id).Scan(&isDisabled)
	if err != nil {
		return err
	}
	if isDisabled {
		return validation.NewFieldError("ContactMethod", "contact method disabled")
	}

	r, err := tx.StmtContext(ctx, s.updateLastSendTime).ExecContext(ctx, id, fmt.Sprintf("%f seconds", minTimeBetweenTests.Seconds()))
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

	vID := uuid.New().String()
	_, err = tx.StmtContext(ctx, s.insertTestNotification).ExecContext(ctx, vID, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) SendContactMethodVerification(ctx context.Context, cmID string) error {
	_, err := s.cmUserID(ctx, cmID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "notification: send verification message", tx)

	r, err := tx.StmtContext(ctx, s.updateLastSendTime).ExecContext(ctx, cmID, fmt.Sprintf("%f seconds", minTimeBetweenTests.Seconds()))
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

	vcID := uuid.New().String()
	code := s.rand.Intn(900000) + 100000
	_, err = tx.StmtContext(ctx, s.setVerificationCode).ExecContext(ctx, vcID, cmID, code)
	if err != nil {
		return errors.Wrap(err, "set verification code")
	}

	return tx.Commit()
}

func (s *Store) VerifyContactMethod(ctx context.Context, cmID string, code int) error {
	_, err := s.cmUserID(ctx, cmID)
	if err != nil {
		return err
	}

	res, err := s.verifyAndEnableContactMethod.ExecContext(ctx, cmID, code)
	if errors.Is(err, sql.ErrNoRows) {
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

func messageStateFromStatus(lastStatus string, hasNextRetry bool) (State, error) {
	switch lastStatus {
	case "queued_remotely", "sending":
		return StateSending, nil
	case "pending":
		return StatePending, nil
	case "sent":
		return StateSent, nil
	case "delivered":
		return StateDelivered, nil
	case "failed", "bundled": // bundled message was not sent (replaced) and should never be re-sent
		// temporary if retry
		if hasNextRetry {
			return StateFailedTemp, nil
		} else {
			return StateFailedPerm, nil
		}
	default:
		return -1, fmt.Errorf("unknown last_status %s", lastStatus)
	}
}

func (s *Store) FindManyMessageStatuses(ctx context.Context, strIDs []string) ([]SendResult, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if len(strIDs) == 0 {
		return nil, nil
	}

	ids, err := validate.ParseManyUUID("IDs", strIDs, search.MaxResults)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(s.db).NfyManyMessageStatus(ctx, ids)
	if err != nil {
		return nil, err
	}

	var result []SendResult
	for _, r := range rows {
		res, err := outgoingMessageToSendResult(r.OutgoingMessage, r.CmDest, r.ChDest)
		if err != nil {
			return nil, err
		}
		result = append(result, *res)
	}

	return result, nil
}

// LastMessageStatus will return the MessageStatus and creation time of the most recent message of the requested type
// for the provided contact method ID, if one was created from the provided from time.
func (s *Store) LastMessageStatus(ctx context.Context, typ MessageType, cmIDStr string, from time.Time) (*SendResult, time.Time, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, time.Time{}, err
	}

	cmID, err := validate.ParseUUID("Contact Method ID", cmIDStr)
	if err != nil {
		return nil, time.Time{}, err
	}

	row, err := gadb.New(s.db).NfyLastMessageStatus(ctx, gadb.NfyLastMessageStatusParams{
		MessageType:     gadb.EnumOutgoingMessagesType(typ),
		ContactMethodID: uuid.NullUUID{UUID: cmID, Valid: true},
		CreatedAt:       from,
	})
	if err != nil {
		return nil, time.Time{}, err
	}

	sendRes, err := outgoingMessageToSendResult(row.OutgoingMessage, row.CmDest, row.ChDest)
	if err != nil {
		return nil, time.Time{}, err
	}

	return sendRes, row.OutgoingMessage.CreatedAt, nil
}
