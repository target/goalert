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

	findMany *sql.Stmt

	deleteOne          *sql.Stmt
	userRotations      *sql.Stmt
	rotationParts      *sql.Stmt
	updateRotationPart *sql.Stmt
	deleteRotationPart *sql.Stmt
	rotActiveIndex     *sql.Stmt
	rotSetActive       *sql.Stmt
	lockRotTables      *sql.Stmt

	findOneForUpdate *sql.Stmt

	findOneBySubject *sql.Stmt

	insertUserAuthSubject *sql.Stmt
	deleteUserAuthSubject *sql.Stmt

	usersMissingProvider *sql.Stmt
	setAuthSubject       *sql.Stmt

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

		insert: p.P(`
			INSERT INTO users (
				id, name, email, avatar_url, role
			)
			VALUES ($1, $2, $3, $4, $5)
		`),

		ids: p.P(`SELECT id FROM users`),

		update: p.P(`
			UPDATE users
			SET
				name = $2,
				email = $3
			WHERE id = $1
		`),

		rotActiveIndex: p.P(`SELECT position FROM rotation_state WHERE rotation_id = $1 FOR UPDATE`),
		rotSetActive:   p.P(`UPDATE rotation_state SET position = $2, rotation_participant_id = $3 WHERE rotation_id = $1`),
		lockRotTables:  p.P(`LOCK TABLE rotation_participants, rotation_state IN EXCLUSIVE MODE`),

		setUserRole: p.P(`UPDATE users SET role = $2 WHERE id = $1`),
		findAuthSubjects: p.P(`
			select subject_id, user_id, provider_id
			from auth_subjects
			where
				(provider_id = $1 or $1 isnull) and
				(user_id = any($2) or $2 isnull)
		`),

		usersMissingProvider: p.P(`
			SELECT
				id, name, email, avatar_url, role, false
			FROM users
			WHERE id not in (select user_id from auth_subjects where provider_id = $1)
		`),
		setAuthSubject: p.P(`
			INSERT INTO auth_subjects (provider_id, subject_id, user_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (provider_id, subject_id) DO UPDATE
			SET user_id = $3
		`),

		findMany: p.P(`
			SELECT
				u.id, u.name, u.email, u.avatar_url, u.role, fav is distinct from null
			FROM users u
			LEFT JOIN user_favorites fav ON
				fav.tgt_user_id = u.id AND fav.user_id = $2
			WHERE u.id = any($1)
		`),

		deleteOne:          p.P(`DELETE FROM users WHERE id = $1`),
		userRotations:      p.P(`SELECT DISTINCT rotation_id FROM rotation_participants WHERE user_id = $1`),
		rotationParts:      p.P(`SELECT id, user_id FROM rotation_participants WHERE rotation_id = $1 ORDER BY position`),
		updateRotationPart: p.P(`UPDATE rotation_participants SET user_id = $2 WHERE id = $1`),
		deleteRotationPart: p.P(`DELETE FROM rotation_participants WHERE id = $1`),

		findOneBySubject: p.P(`
			SELECT
				u.id, u.name, u.email, u.avatar_url, u.role, false
			FROM auth_subjects s
			JOIN users u ON u.id = s.user_id
			WHERE s.provider_id = $1 AND s.subject_id = $2
		`),

		findOne: p.P(`
			SELECT
				u.id, u.name, u.email, u.avatar_url, u.role, fav is distinct from null
			FROM users u
			LEFT JOIN user_favorites fav ON
				fav.tgt_user_id = u.id AND fav.user_id = $2
			WHERE u.id = $1
		`),

		findOneForUpdate: p.P(`
			SELECT
				id, name, email, avatar_url, role, false
			FROM users
			WHERE id = $1
			FOR UPDATE
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

// SetAuthSubject will add or update the auth subject for the provider/subject pair to point to the provided user ID.
func (s *Store) SetAuthSubject(ctx context.Context, providerID, subjectID, userID string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.SubjectID("ProviderID", providerID),
		validate.SubjectID("SubjectID", subjectID),
		validate.UUID("UserID", userID),
	)
	if err != nil {
		return err
	}

	_, err = s.setAuthSubject.ExecContext(ctx, providerID, subjectID, userID)
	if err != nil {
		return err
	}

	return nil
}

// WithoutAuthProviderFunc will call forEachFn for each user that is missing an auth subject for the given provider ID.
// If an error is returned by forEachFn it will stop reading and be returned. Favorites information will not be included (always false).
func (s *Store) WithoutAuthProviderFunc(ctx context.Context, providerID string, forEachFn func(User) error) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}
	err = validate.SubjectID("ProviderID", providerID)
	if err != nil {
		return err
	}

	rows, err := s.usersMissingProvider.QueryContext(ctx, providerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var u User
		err = u.scanFrom(rows.Scan)
		if err != nil {
			return fmt.Errorf("scan user row with missing provider '%s': %w", providerID, err)
		}
		err = forEachFn(u)
		if err != nil {
			return err
		}
	}

	return nil
}

// AuthSubjectsFunc will call the provided forEachFn for each AuthSubject.
// If an error is returned by forEachFn it will stop reading subjects and be returned.
//
// providerID, if not empty, will limit AuthSubjects to those with the same providerID.
// userID, if not empty, will limit AuthSubjects to those assigned to the given userID(s).
func (s *Store) AuthSubjectsFunc(ctx context.Context, providerID string, userIDs []string, forEachFn func(AuthSubject) error) error {
	checks := []permission.Checker{permission.Admin}
	if len(userIDs) == 1 {
		checks = append(checks, permission.MatchUser(userIDs[0]))
	}
	err := permission.LimitCheckAny(ctx, checks...)
	if err != nil {
		return err
	}
	if providerID != "" {
		err = validate.SubjectID("ProviderID", providerID)
	}
	err = validate.Many(err, validate.ManyUUID("UserID", userIDs, 100))
	if err != nil {
		return err
	}

	pID := sql.NullString{
		String: providerID,
		Valid:  providerID != "",
	}

	var uIDs sqlutil.NullUUIDArray
	if len(userIDs) > 0 {
		uIDs.Valid = true
		uIDs.UUIDArray = sqlutil.UUIDArray(userIDs)
	}

	rows, err := s.findAuthSubjects.QueryContext(ctx, pID, uIDs)
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
		defer sqlutil.Rollback(ctx, "user: delete", tx)
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
	_, err := tx.StmtContext(ctx, s.lockRotTables).ExecContext(ctx)
	if err != nil {
		return err
	}

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

	var activeIndex int
	err = tx.StmtContext(ctx, s.rotActiveIndex).QueryRowContext(ctx, rotationID).Scan(&activeIndex)
	if err != nil {
		return fmt.Errorf("query active index: %w", err)
	}

	// update participant user IDs
	var skipped bool
	curIndex := -1
	updatePart := tx.StmtContext(ctx, s.updateRotationPart)
	for i, p := range participants {
		if p.UserID == userID {
			if i < activeIndex {
				activeIndex--
			}

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
	if activeIndex > curIndex {
		activeIndex = 0
	}

	// delete in reverse order from the end
	deletePart := tx.StmtContext(ctx, s.deleteRotationPart)
	for i := len(participants) - 1; i > curIndex; i-- {
		_, err = deletePart.ExecContext(ctx, participants[i].ID)
		if err != nil {
			return fmt.Errorf("delete participant %d: %w", i, err)
		}
	}

	_, err = tx.StmtContext(ctx, s.rotSetActive).ExecContext(ctx, rotationID, activeIndex, participants[activeIndex].ID)
	if err != nil {
		return fmt.Errorf("set active index: %w", err)
	}

	return nil
}

// UpdateTx allows updating a user name and email.
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

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids), ctxFavIDParam(ctx))
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

func ctxFavIDParam(ctx context.Context) sql.NullString {
	userID := permission.UserID(ctx)
	if userID == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: userID, Valid: true}
}

// FindOneTx will return a single user, locking the row if forUpdate is set. When `forUpdate` is true,
// favorite information is omitted (always false).
func (s *Store) FindOneTx(ctx context.Context, tx *sql.Tx, id string, forUpdate bool) (*User, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("UserID", id)
	if err != nil {
		return nil, err
	}

	var row *sql.Row
	if forUpdate {
		row = withTx(ctx, tx, s.findOneForUpdate).QueryRowContext(ctx, id)
	} else {
		row = withTx(ctx, tx, s.findOne).QueryRowContext(ctx, id, ctxFavIDParam(ctx))
	}

	var u User
	err = u.scanFrom(row.Scan)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindOneBySubject will find a user matching the subjectID for the given providerID.
func (s *Store) FindOneBySubject(ctx context.Context, providerID, subjectID string) (*User, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.Many(
		validate.SubjectID("ProviderID", providerID),
		validate.SubjectID("SubjectID", subjectID),
	)
	if err != nil {
		return nil, err
	}

	row := s.findOneBySubject.QueryRowContext(ctx, providerID, subjectID)
	var u User
	err = u.scanFrom(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
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
	err := s.AuthSubjectsFunc(ctx, "", []string{userID}, func(sub AuthSubject) error {
		result = append(result, sub)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
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
