package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/golang/groupcache"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// Store allows the lookup and management of Users.
type Store interface {
	Insert(context.Context, *User) (*User, error)
	InsertTx(context.Context, *sql.Tx, *User) (*User, error)
	Update(context.Context, *User) error
	UpdateTx(context.Context, *sql.Tx, *User) error
	Delete(context.Context, string) error
	DeleteManyTx(context.Context, *sql.Tx, []string) error
	FindOne(context.Context, string) (*User, error)
	FindOneTx(ctx context.Context, tx *sql.Tx, id string, forUpdate bool) (*User, error)
	FindAll(context.Context) ([]User, error)
	FindMany(context.Context, []string) ([]User, error)
	Search(context.Context, *SearchOptions) ([]User, error)

	UserExists(context.Context) (ExistanceChecker, error)

	AddAuthSubjectTx(ctx context.Context, tx *sql.Tx, a *AuthSubject) error
	DeleteAuthSubjectTx(ctx context.Context, tx *sql.Tx, a *AuthSubject) error
	FindAllAuthSubjectsForUser(ctx context.Context, userID string) ([]AuthSubject, error)
	StreamAuthSubjects(ctx context.Context, providerID, userID string, eachFn func(AuthSubject) error) error
	FindSomeAuthSubjectsForProvider(ctx context.Context, limit int, afterSubjectID, providerID string) ([]AuthSubject, error)
}

var _ Store = &DB{}

// DB implements the Store against a *sql.DB backend.
type DB struct {
	db *sql.DB

	ids *sql.Stmt

	insert  *sql.Stmt
	update  *sql.Stmt
	delete  *sql.Stmt
	findOne *sql.Stmt
	findAll *sql.Stmt

	findMany   *sql.Stmt
	deleteMany *sql.Stmt

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

// NewDB will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	store := &DB{
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

		delete: p.P(`
			DELETE FROM users
			WHERE id = $1
		`),

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
		deleteMany: p.P(`DELETE FROM users WHERE id = any($1)`),

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

func (db *DB) StreamAuthSubjects(ctx context.Context, providerID, userID string, forEachFn func(AuthSubject) error) error {
	err := permission.LimitCheckAny(ctx, permission.System)
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

	rows, err := db.findAuthSubjects.QueryContext(ctx, pID, uID)
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

func (db *DB) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
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

	del := db.deleteMany
	if tx != nil {
		tx.StmtContext(ctx, del)
	}

	_, err = del.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

// InsertTx implements the Store interface by inserting the new User into the database.
// The insert statement is first wrapped in tx.
func (db *DB) InsertTx(ctx context.Context, tx *sql.Tx, u *User) (*User, error) {
	n, err := u.Normalize()
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}
	_, err = tx.Stmt(db.insert).ExecContext(ctx, n.fields()...)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Insert implements the Store interface by inserting the new User into the database.
func (db *DB) Insert(ctx context.Context, u *User) (*User, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	u, err = db.InsertTx(ctx, tx, u)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return u, nil
}

// Delete implements the UserStore interface.
func (db *DB) Delete(ctx context.Context, id string) error {
	err := validate.UUID("UserID", id)
	if err != nil {
		return err
	}
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}
	_, err = db.delete.ExecContext(ctx, id)
	return err
}

// Update implements the Store interface. Only admins can update user roles.
func (db *DB) Update(ctx context.Context, u *User) error {
	return db.UpdateTx(ctx, nil, u)
}
func (db *DB) UpdateTx(ctx context.Context, tx *sql.Tx, u *User) error {
	n, err := u.Normalize()
	if err != nil {
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(u.ID))
	if err != nil {
		return err
	}
	update := db.update
	if tx != nil {
		update = tx.StmtContext(ctx, update)
	}
	_, err = update.ExecContext(ctx, n.userUpdateFields()...)
	return err
}

func (db *DB) FindMany(ctx context.Context, ids []string) ([]User, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.ManyUUID("UserID", ids, 200)
	if err != nil {
		return nil, err
	}

	rows, err := db.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids))
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

// FindOne implements the Store interface.
func (db *DB) FindOne(ctx context.Context, id string) (*User, error) {
	return db.FindOneTx(ctx, nil, id, false)
}
func (db *DB) FindOneTx(ctx context.Context, tx *sql.Tx, id string, forUpdate bool) (*User, error) {
	err := validate.UUID("UserID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	var u User
	findOne := db.findOne
	if forUpdate {
		findOne = db.findOneForUpdate
	}
	if tx != nil {
		findOne = tx.StmtContext(ctx, findOne)
	}
	row := findOne.QueryRowContext(ctx, id)
	err = u.scanFrom(row.Scan)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindSomeAuthSubjectsForProvider implements the Store interface. It finds all auth subjects associated with a given userID.
func (db *DB) FindSomeAuthSubjectsForProvider(ctx context.Context, limit int, afterSubjectID, providerID string) ([]AuthSubject, error) {
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

	rows, err := db.db.QueryContext(ctx, q, providerID, afterSubjectID)
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

// FindAllAuthSubjectsForUser implements the Store interface. It finds all auth subjects associated with a given userID.
func (db *DB) FindAllAuthSubjectsForUser(ctx context.Context, userID string) ([]AuthSubject, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	var authSubjects []AuthSubject
	rows, err := db.findAuthSubjectsByUser.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a AuthSubject
		a.UserID = userID
		err = rows.Scan(&a.ProviderID, &a.SubjectID)
		if err != nil {
			return nil, err
		}
		authSubjects = append(authSubjects, a)
	}

	return authSubjects, nil
}

// FindAll implements the Store interface.
func (db *DB) FindAll(ctx context.Context) ([]User, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	rows, err := db.findAll.QueryContext(ctx)
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

// AddAuthSubjectTx implements the Store interface. It is used to add an auth subject to a given user.
func (db *DB) AddAuthSubjectTx(ctx context.Context, tx *sql.Tx, a *AuthSubject) error {
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

	s := db.insertUserAuthSubject
	if tx != nil {
		s = tx.Stmt(s)
	}
	_, err = s.ExecContext(ctx, a.UserID, n.ProviderID, n.SubjectID)
	return err
}

// DeleteAuthSubjectTx implements the Store interface. It is used to remove an auth subject for a given user.
func (db *DB) DeleteAuthSubjectTx(ctx context.Context, tx *sql.Tx, a *AuthSubject) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}

	n, err := a.Normalize()
	if err != nil {
		return err
	}

	s := db.deleteUserAuthSubject
	if tx != nil {
		s = tx.Stmt(s)
	}
	_, err = s.ExecContext(ctx, a.UserID, n.ProviderID, n.SubjectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// do not return error if auth subject doesn't exist
		return err
	}
	return nil
}
