package service

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

type Store interface {
	FindMany(context.Context, []string) ([]Service, error)

	FindOne(context.Context, string) (*Service, error)
	FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Service, error)
	FindOneForUser(ctx context.Context, userID, serviceID string) (*Service, error)
	FindAll(context.Context) ([]Service, error)
	DeleteManyTx(context.Context, *sql.Tx, []string) error
	Insert(context.Context, *Service) (*Service, error)
	CreateServiceTx(context.Context, *sql.Tx, *Service) (*Service, error)

	Update(context.Context, *Service) error
	UpdateTx(context.Context, *sql.Tx, *Service) error

	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error

	FindAllByEP(context.Context, string) ([]Service, error)
	LegacySearch(ctx context.Context, opts *LegacySearchOptions) ([]Service, error)
	Search(ctx context.Context, opts *SearchOptions) ([]Service, error)
}

type DB struct {
	db *sql.DB

	findOne     *sql.Stmt
	findOneUp   *sql.Stmt
	findMany    *sql.Stmt
	findAll     *sql.Stmt
	findAllByEP *sql.Stmt
	insert      *sql.Stmt
	update      *sql.Stmt
	delete      *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}
	p := prep.P

	s := &DB{db: db}
	s.findOne = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			fav	is distinct from null
		FROM
			services s
		JOIN escalation_policies e ON e.id = s.escalation_policy_id
		LEFT JOIN user_favorites fav ON s.id = fav.tgt_service_id AND fav.user_id = $2
		WHERE
			s.id = $1
	`)
	s.findOneUp = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id
		FROM services s
		WHERE s.id = $1
		FOR UPDATE
	`)
	s.findMany = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			fav	is distinct from null
		FROM
			services s
		JOIN escalation_policies e ON e.id = s.escalation_policy_id
		LEFT JOIN user_favorites fav ON s.id = fav.tgt_service_id AND fav.user_id = $2
		WHERE
			s.id = any($1)
	`)

	s.findAll = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			false
		FROM
			services s,
			escalation_policies e
		WHERE
			e.id = s.escalation_policy_id
	`)
	s.findAllByEP = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			false
		FROM
			services s,
			escalation_policies e
		WHERE
			e.id = $1 AND
			e.id = s.escalation_policy_id
	`)
	s.insert = p(`INSERT INTO services (id,name,description,escalation_policy_id) VALUES ($1,$2,$3,$4)`)
	s.update = p(`UPDATE services SET name = $2, description = $3, escalation_policy_id = $4 WHERE id = $1`)
	s.delete = p(`DELETE FROM services WHERE id = any($1)`)

	return s, prep.Err
}

func (db *DB) FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Service, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", id)
	if err != nil {
		return nil, err
	}
	var s Service
	err = tx.StmtContext(ctx, db.findOneUp).QueryRowContext(ctx, id).Scan(&s.ID, &s.Name, &s.Description, &s.EscalationPolicyID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindMany returns slice of Service objects given a slice of serviceIDs
func (db *DB) FindMany(ctx context.Context, ids []string) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	err = validate.ManyUUID("ServiceIDs", ids, 100)
	if err != nil {
		return nil, err
	}

	rows, err := db.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids), permission.UserID(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}

func (db *DB) Insert(ctx context.Context, s *Service) (*Service, error) {
	return db.CreateServiceTx(ctx, nil, s)
}
func (db *DB) CreateServiceTx(ctx context.Context, tx *sql.Tx, s *Service) (*Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := s.Normalize()
	if err != nil {
		return nil, err
	}

	n.ID = uuid.NewV4().String()
	stmt := db.insert
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.EscalationPolicyID)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Delete implements the ServiceInterface interface.
func (db *DB) Delete(ctx context.Context, id string) error {
	return db.DeleteTx(ctx, nil, id)
}
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	return db.DeleteManyTx(ctx, tx, []string{id})
}
func (db *DB) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("ServiceID", ids, 50)
	if err != nil {
		return err
	}
	s := db.delete
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}
	_, err = s.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func wrap(tx *sql.Tx, s *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return s
	}
	return tx.Stmt(s)
}

// Update implements the ServiceStore interface.
func (db *DB) Update(ctx context.Context, s *Service) error {
	return db.UpdateTx(ctx, nil, s)
}
func (db *DB) UpdateTx(ctx context.Context, tx *sql.Tx, s *Service) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	n, err := s.Normalize()
	if err != nil {
		return err
	}

	err = validate.UUID("ServiceID", n.ID)
	if err != nil {
		return err
	}

	_, err = wrap(tx, db.update).ExecContext(ctx, n.ID, n.Name, n.Description, n.EscalationPolicyID)
	return err
}

func (db *DB) FindOneForUser(ctx context.Context, userID, serviceID string) (*Service, error) {
	err := validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	var uid sql.NullString
	userCheck := permission.User

	if userID != "" {
		err := validate.UUID("UserID", userID)
		if err != nil {
			return nil, err
		}
		userCheck = permission.MatchUser(userID)
		uid.Valid = true
		uid.String = userID
	}

	err = permission.LimitCheckAny(ctx, userCheck, permission.System)
	if err != nil {
		return nil, err
	}

	row := db.findOne.QueryRowContext(ctx, serviceID, uid)
	var s Service
	err = scanFrom(&s, row.Scan)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (db *DB) FindOne(ctx context.Context, id string) (*Service, error) {
	// old method just calls new method
	return db.FindOneForUser(ctx, "", id)
}

func scanFrom(s *Service, f func(args ...interface{}) error) error {
	return f(&s.ID, &s.Name, &s.Description, &s.EscalationPolicyID, &s.epName, &s.isUserFavorite)
}

func scanAllFrom(rows *sql.Rows) (services []Service, err error) {
	var s Service
	for rows.Next() {
		err = scanFrom(&s, rows.Scan)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (db *DB) FindAll(ctx context.Context) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}
func (db *DB) FindAllByEP(ctx context.Context, epID string) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAllByEP.QueryContext(ctx, epID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}
