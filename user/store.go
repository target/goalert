package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache"
	"github.com/google/uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// Store allows managing users.
type Store struct {
	db *sql.DB

	ids *sql.Stmt

	insert      *sql.Stmt
	update      *sql.Stmt
	setUserRole *sql.Stmt
	findOne     *sql.Stmt
	findAll     *sql.Stmt

	findMany *sql.Stmt

	deleteOne          *sql.Stmt
	userRotations      *sql.Stmt
	rotationParts      *sql.Stmt
	updateRotationPart *sql.Stmt
	deleteRotationPart *sql.Stmt

	findOneForUpdate *sql.Stmt

	insertUserAuthSubject *sql.Stmt
	deleteUserAuthSubject *sql.Stmt

	findAuthSubjectsByUser *sql.Stmt

	findAuthSubjects *sql.Stmt

	grp *groupcache.Group

	userExistHash []byte
	userExist     chan map[uuid.UUID]struct{}
}

var grpN int64

// NewStore will create new Store for the sql.DB. An error will be returned if statements fail to prepare.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	store := &Store{
		db: db,

		userExist: make(chan map[uuid.UUID]struct{}, 1),

		ids: p.P(`SELECT id FROM users`),

		insert: p.P(`
			INSERT INTO users (
				id, name, email, avatar_url, role, alert_status_log_contact_method_id
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`),

		update: p.P(`
			UPDATE users
			SET
				name = $2,
				email = $3,
				alert_status_log_contact_method_id = $4
			WHERE id = $1
		`),

		setUserRole: p.P(`UPDATE users SET role = $2 WHERE id = $1`),
		findAuthSubjects: p.P(`
			select subject_id, user_id, provider_id
			from auth_subjects
			where
				(provider_id = $1 or $1 isnull) and
				(user_id = $2 or $2 isnull)
		`),

		findMany: p.P(`
			SELECT
				id, name, email, avatar_url, role, alert_status_log_contact_method_id
			FROM users
			WHERE id = any($1)
		`),

		deleteOne:          p.P(`DELETE FROM users WHERE id = $1`),
		userRotations:      p.P(`SELECT DISTINCT rotation_id FROM rotation_participants WHERE user_id = $1`),
		rotationParts:      p.P(`SELECT id, user_id FROM rotation_participants WHERE rotation_id = $1 ORDER BY position`),
		updateRotationPart: p.P(`UPDATE rotation_participants SET user_id = $2 WHERE id = $1`),
		deleteRotationPart: p.P(`DELETE FROM rotation_participants WHERE id = $1`),

		findOne: p.P(`
			SELECT
				id, name, email, avatar_url, role, alert_status_log_contact_method_id
			FROM users
			WHERE id = $1
		`),
		findOneForUpdate: p.P(`
			SELECT
				id, name, email, avatar_url, role, alert_status_log_contact_method_id
			FROM users
			WHERE id = $1
			FOR UPDATE
		`),

		findAuthSubjectsByUser: p.P(`
			SELECT provider_id, subject_id
			FROM auth_subjects 
			WHERE user_id = $1
		`),

		findAll: p.P(`
			SELECT
				id, name, email, avatar_url, role, alert_status_log_contact_method_id
			FROM users
		`),

		insertUserAuthSubject: p.P(`
			INSERT into auth_subjects (
				user_id, provider_id, subject_id
			)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`),

		deleteUserAuthSubject: p.P(`
			DELETE FROM auth_subjects
			WHERE 
				user_id = $1 AND
				provider_id = $2 AND 
				subject_id = $3
		`),
	}
	if p.Err != nil {
		return nil, p.Err
	}

	store.userExist <- make(map[uuid.UUID]struct{})

	store.grp = groupcache.NewGroup(fmt.Sprintf("user.store[%d]", atomic.AddInt64(&grpN, 1)), 1024*1024, groupcache.GetterFunc(store.cacheGet))

	return store, nil
}

// AuthSubjectsFunc will call the provided forEachFn for each auth subject for the given userID and/or providerID.
// If an error is returned by forEachFn it will stop reading subjects and be returned.
//
// If providerID is empty, all providers will be returned.
// if userID is empty, all users will be returned.
func (s *Store) AuthSubjectsFunc(ctx context.Context, providerID, userID string, forEachFn func(AuthSubject) error) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}
	if providerID != "" {
		err = validate.SubjectID("ProviderID", providerID)
	}
	if userID != "" {
		err = validate.Many(err, validate.UUID("UserID", userID))
	}
	if err != nil {
		return err
	}

	pID := sql.NullString{
		String: providerID,
		Valid:  providerID != "",
	}
	uID := sql.NullString{
		String: userID,
		Valid:  userID != "",
	}

	rows, err := s.findAuthSubjects.QueryContext(ctx, pID, uID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sub AuthSubject
		err = rows.Scan(&sub.SubjectID, &sub.UserID, &sub.ProviderID)
		if err != nil {
			return err
		}
		err = forEachFn(sub)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteManyTx will delete multiple users within the same transaction. If tx is nil,
// a transaction will be started and committed before returning.
func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	err = validate.Range("Count", len(ids), 1, 100)
	if err != nil {
		return err
	}

	var ownsTx bool
	if tx == nil {
		ownsTx = true
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	for _, id := range ids {
		err = s.retryDeleteTx(ctx, tx, id)
		if err != nil {
			return err
		}
	}

	if ownsTx {
		return tx.Commit()
	}

	return nil
}

func withTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}
func (s *Store) requireTx(ctx context.Context, tx *sql.Tx, fn func(*sql.Tx) error) error {
	return nil
}

// InsertTx creates a new User.
func (s *Store) InsertTx(ctx context.Context, tx *sql.Tx, u *User) (*User, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}

	n, err := u.Normalize()
	if err != nil {
		return nil, err
	}

	_, err = withTx(ctx, tx, s.insert).ExecContext(ctx, n.fields()...)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Insert is equivalent to calling InsertTx(ctx, nil, u).
func (s *Store) Insert(ctx context.Context, u *User) (*User, error) { return s.InsertTx(ctx, nil, u) }

// DeleteTx deletes a User with the given ID.
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}

	err = validate.UUID("UserID", id)
	if err != nil {
		return err
	}

	return s.requireTx(ctx, tx, func(tx *sql.Tx) error { return s.retryDeleteTx(ctx, tx, id) })
}

func (s *Store) retryDeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	return retry.DoTemporaryError(func(int) error {
		err := s._deleteTx(ctx, tx, id)
		sqlErr := sqlutil.MapError(err)
		if sqlErr != nil && sqlErr.Code == "23503" {
			// retry foreign key errors when deleting a user
			err = retry.TemporaryError(err)
		}
		return err
	},
		retry.Log(ctx),
		retry.Limit(5),
		retry.FibBackoff(250*time.Millisecond),
	)
}

func (s *Store) _deleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	// cleanup rotations first
	rows, err := tx.StmtContext(ctx, s.userRotations).QueryContext(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("lookup user rotations: %w", err)
	}
	defer rows.Close()

	var rotationIDs []string
	for rows.Next() {
		var rID string
		err = rows.Scan(&rID)
		if err != nil {
			return fmt.Errorf("scan user rotation id: %w", err)
		}
		rotationIDs = append(rotationIDs, rID)
	}

	for _, rID := range rotationIDs {
		err = s.removeUserFromRotation(ctx, tx, id, rID)
		if err != nil {
			return fmt.Errorf("remove user '%s' from rotation '%s': %w", id, rID, err)
		}
	}

	_, err = tx.StmtContext(ctx, s.deleteOne).ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user row: %w", err)
	}
	return nil
}

func (s *Store) removeUserFromRotation(ctx context.Context, tx *sql.Tx, userID, rotationID string) error {
	type part struct {
		ID     string
		UserID string
	}
	var participants []part
	rows, err := tx.StmtContext(ctx, s.rotationParts).QueryContext(ctx, rotationID)
	if err != nil {
		return fmt.Errorf("query participants: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p part
		err = rows.Scan(&p.ID, &p.UserID)
		if err != nil {
			return fmt.Errorf("scan participant %d: %w", len(participants), err)
		}
		participants = append(participants, p)
	}

	// update participant user IDs
	var skipped bool
	curIndex := -1
	updatePart := tx.StmtContext(ctx, s.updateRotationPart)
	for _, p := range participants {
		if p.UserID == userID {
			skipped = true
			continue
		}
		curIndex++
		if skipped {
			_, err = updatePart.ExecContext(ctx, participants[curIndex].ID, p.UserID)
			if err != nil {
				return fmt.Errorf("update participant %d to user '%s': %w", curIndex, p.UserID, err)
			}
		}
	}

	// delete in reverse order from the end
	deletePart := tx.StmtContext(ctx, s.deleteRotationPart)
	for i := len(participants) - 1; i > curIndex; i-- {
		_, err = deletePart.ExecContext(ctx, participants[i].ID)
		if err != nil {
			return fmt.Errorf("delete participant %d: %w", i, err)
		}
	}

	return nil
}

// Delete is equivalent to calling DeleteTx(ctx, nil, id).
func (s *Store) Delete(ctx context.Context, id string) error { return s.DeleteTx(ctx, nil, id) }

// Update id equivalent to calling UpdateTx(ctx, nil, u).
func (s *Store) Update(ctx context.Context, u *User) error { return s.UpdateTx(ctx, nil, u) }

// UpdateTx allows updating a user name, email, and status update preference.
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, u *User) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(u.ID))
	if err != nil {
		return err
	}

	n, err := u.Normalize()
	if err != nil {
		return err
	}

	_, err = withTx(ctx, tx, s.update).ExecContext(ctx, n.userUpdateFields()...)
	return err
}

// SetUserRoleTx allows updating the role of the given user ID.
func (s *Store) SetUserRoleTx(ctx context.Context, tx *sql.Tx, id string, role permission.Role) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("UserID", id),
		validate.OneOf("Role", role, permission.RoleAdmin, permission.RoleUser),
	)
	if err != nil {
		return err
	}

	_, err = withTx(ctx, tx, s.setUserRole).ExecContext(ctx, id, role)
	return err
}

// FindMany will return all users matching the provided IDs.
//
// There is no guarantee the returned users will be in the same order or
// that the number of returned users matches the number of provided IDs.
func (s *Store) FindMany(ctx context.Context, ids []string) ([]User, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.ManyUUID("UserID", ids, 200)
	if err != nil {
		return nil, err
	}

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]User, 0, len(ids))
	var u User
	for rows.Next() {
		err = u.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

// FindOne is equivalent to calling FindOneTx(ctx, nil, id, false).
func (s *Store) FindOne(ctx context.Context, id string) (*User, error) {
	return s.FindOneTx(ctx, nil, id, false)
}

// FindOneTx will return a single user, locking the row if forUpdate is set.
func (s *Store) FindOneTx(ctx context.Context, tx *sql.Tx, id string, forUpdate bool) (*User, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("UserID", id)
	if err != nil {
		return nil, err
	}

	stmt := s.findOne
	if forUpdate {
		stmt = s.findOneForUpdate
	}
	row := withTx(ctx, tx, stmt).QueryRowContext(ctx, id)
	var u User
	err = u.scanFrom(row.Scan)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindSomeAuthSubjectsForProvider returns up to `limit` auth subjects associated with a given providerID.
//
// afterSubjectID can be specified for paginating responses. Results are sorted by subject id.
func (s *Store) FindSomeAuthSubjectsForProvider(ctx context.Context, limit int, afterSubjectID, providerID string) ([]AuthSubject, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}

	// treat as a subject ID for now
	err = validate.Many(
		validate.SubjectID("ProviderID", providerID),
		validate.Range("Limit", limit, 0, 9000),
	)
	if afterSubjectID != "" {
		err = validate.Many(err, validate.SubjectID("AfterID", afterSubjectID))
	}
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = 50
	}

	q := fmt.Sprintf(`
		SELECT user_id, subject_id
		FROM auth_subjects
		WHERE provider_id = $1 AND subject_id > $2
		ORDER BY subject_id
		LIMIT %d
	`, limit)

	rows, err := s.db.QueryContext(ctx, q, providerID, afterSubjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var authSubjects []AuthSubject
	for rows.Next() {
		var a AuthSubject
		a.ProviderID = providerID

		err = rows.Scan(&a.UserID, &a.SubjectID)
		if err != nil {
			return nil, err
		}
		authSubjects = append(authSubjects, a)
	}

	return authSubjects, nil
}

// FindAllAuthSubjectsForUser returns all auth subjects associated with a given userID.
func (s *Store) FindAllAuthSubjectsForUser(ctx context.Context, userID string) ([]AuthSubject, error) {
	var result []AuthSubject
	err := s.AuthSubjectsFunc(ctx, "", userID, func(sub AuthSubject) error {
		result = append(result, sub)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindAll returns all users.
func (s *Store) FindAll(ctx context.Context) ([]User, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	rows, err := s.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		err = u.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// AddAuthSubjectTx adds an auth subject for a user.
func (s *Store) AddAuthSubjectTx(ctx context.Context, tx *sql.Tx, a *AuthSubject) error {
	var userID string
	if a != nil {
		userID = a.UserID
	}
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	n, err := a.Normalize()
	if err != nil {
		return err
	}

	_, err = withTx(ctx, tx, s.insertUserAuthSubject).ExecContext(ctx, a.UserID, n.ProviderID, n.SubjectID)
	return err
}

// DeleteAuthSubjectTx removes an auth subject for a user.
//
// If the subject does not exist, nil is returned.
func (s *Store) DeleteAuthSubjectTx(ctx context.Context, tx *sql.Tx, a *AuthSubject) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}

	n, err := a.Normalize()
	if err != nil {
		return err
	}

	_, err = withTx(ctx, tx, s.deleteUserAuthSubject).ExecContext(ctx, a.UserID, n.ProviderID, n.SubjectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// do not return error if auth subject doesn't exist
		return err
	}
	return nil
}
