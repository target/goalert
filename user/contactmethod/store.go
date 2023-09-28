package contactmethod

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Store implements the lookup and management of ContactMethods against a *sql.Store backend.
type Store struct {
	db *sql.DB
}

// NewStore will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewStore(ctx context.Context, db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) queries(dbtx gadb.DBTX) *gadb.Queries {
	if dbtx != nil {
		return gadb.New(dbtx)
	}
	return gadb.New(s.db)
}

func (s *Store) MetadataByTypeValue(ctx context.Context, dbtx gadb.DBTX, typ Type, value string) (*Metadata, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	data, err := s.queries(dbtx).MetaTVContactMethod(ctx, gadb.MetaTVContactMethodParams{Type: gadb.EnumUserContactMethodType(typ), Value: value})
	if err != nil {
		return nil, err
	}

	var m Metadata
	err = json.Unmarshal(data.Metadata, &m)
	if err != nil {
		return nil, err
	}
	m.FetchedAt = data.Now

	return &m, nil
}

func (s *Store) SetCarrierV1MetadataByTypeValue(ctx context.Context, tx *sql.Tx, typ Type, value string, newM *Metadata) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	var ownTx bool
	if tx == nil {
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer sqlutil.Rollback(ctx, "cm: set carrier metadata", tx)

		ownTx = true
	}
	m, err := s.MetadataByTypeValue(ctx, tx, typ, value)
	if err != nil {
		return err
	}
	m.CarrierV1 = newM.CarrierV1
	m.CarrierV1.UpdatedAt = m.FetchedAt

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = gadb.New(tx).UpdateMetaTVContactMethod(ctx, gadb.UpdateMetaTVContactMethodParams{Type: gadb.EnumUserContactMethodType(typ), Value: value, Metadata: pqtype.NullRawMessage{RawMessage: data, Valid: data != nil}})
	if err != nil {
		return err
	}

	if ownTx {
		return tx.Commit()
	}

	return nil
}

func (s *Store) EnableByValue(ctx context.Context, t Type, v string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	c := ContactMethod{Name: "Enable", Type: t, Value: v}
	n, err := c.Normalize()
	if err != nil {
		return err
	}

	id, err := gadb.New(s.db).EnableContactMethod(ctx, gadb.EnableContactMethodParams{Type: gadb.EnumUserContactMethodType(n.Type), Value: n.Value})

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method START code received.")
	}

	return err
}

func (s *Store) DisableByValue(ctx context.Context, t Type, v string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	c := ContactMethod{Name: "Disable", Type: t, Value: v}
	n, err := c.Normalize()
	if err != nil {
		return err
	}

	id, err := gadb.New(s.db).DisableContactMethod(ctx, gadb.DisableContactMethodParams{Type: gadb.EnumUserContactMethodType(n.Type), Value: v})

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method STOP code received.")
	}

	return err
}

// CreateTx inserts the new ContactMethod into the database. A new ID is always created.
func (s *Store) Create(ctx context.Context, dbtx gadb.DBTX, c *ContactMethod) (*ContactMethod, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(c.UserID))
	if err != nil {
		return nil, err
	}

	n, err := c.Normalize()
	if err != nil {
		return nil, err
	}

	err = gadb.New(dbtx).AddContactMethod(ctx, gadb.AddContactMethodParams{
		ID:                  uuid.MustParse(n.ID),
		Name:                n.Name,
		Type:                gadb.EnumUserContactMethodType(n.Type),
		Value:               n.Value,
		Disabled:            n.Disabled,
		UserID:              uuid.MustParse(n.UserID),
		EnableStatusUpdates: n.StatusUpdates,
	})
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Delete removes the ContactMethod from the database using the provided ID within a transaction.
func (s *Store) Delete(ctx context.Context, dbtx gadb.DBTX, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	uids, err := validate.ParseManyUUID("ContactMethodID", ids, 50)
	if err != nil {
		return err
	}

	if permission.Admin(ctx) {
		err = gadb.New(dbtx).DeleteContactMethod(ctx, uids)
		return err
	}

	rows, err := gadb.New(s.db).LookupUserIDContactMethod(ctx, uids)
	if err != nil {
		return err
	}

	var checks []permission.Checker
	for _, id := range rows {
		checks = append(checks, permission.MatchUser(id.String()))
	}

	err = permission.LimitCheckAny(ctx, checks...)
	if err != nil {
		return err
	}

	err = gadb.New(dbtx).DeleteContactMethod(ctx, uids)
	return err
}

// FindOneTx finds the contact method from the database using the provided ID within a transaction.
func (s *Store) FindOne(ctx context.Context, dbtx gadb.DBTX, id string) (*ContactMethod, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	methodUUID, err := validate.ParseUUID("ContactMethodID", id)
	if err != nil {
		return nil, err
	}

	row, err := s.queries(dbtx).FindOneUpdateContactMethod(ctx, methodUUID)
	if err != nil {
		return nil, err
	}

	c := ContactMethod{
		ID:               row.ID.String(),
		Name:             row.Name,
		Type:             Type(row.Type),
		Value:            row.Value,
		Disabled:         row.Disabled,
		UserID:           row.UserID.String(),
		Pending:          row.Pending,
		StatusUpdates:    row.EnableStatusUpdates,
		lastTestVerifyAt: row.LastTestVerifyAt,
	}

	return &c, nil
}

// // FindOne finds the contact method from the database using the provided ID.
// func (s *Store) FindOne(ctx context.Context, id string) (*ContactMethod, error) {
// 	methodUUID, err := validate.ParseUUID("ContactMethodID", id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = permission.LimitCheckAny(ctx, permission.All)
// 	if err != nil {
// 		return nil, err
// 	}

// 	row, err := gadb.New(s.db).FindOneContactMethod(ctx, methodUUID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	c := ContactMethod{
// 		ID:               row.ID.String(),
// 		Name:             row.Name,
// 		Type:             Type(row.Type),
// 		Value:            row.Value,
// 		Disabled:         row.Disabled,
// 		UserID:           row.UserID.String(),
// 		Pending:          row.Pending,
// 		StatusUpdates:    row.EnableStatusUpdates,
// 		lastTestVerifyAt: row.LastTestVerifyAt,
// 	}

// 	return &c, nil
// }

// UpdateTx updates the contact method with the newly provided values within a transaction.
func (s *Store) Update(ctx context.Context, dbtx gadb.DBTX, c *ContactMethod) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	n, err := c.Normalize()
	if err != nil {
		return err
	}

	cm, err := s.FindOne(ctx, dbtx, c.ID)
	if err != nil {
		return err
	}
	if n.Type != cm.Type {
		return validation.NewFieldError("Type", "cannot update type of contact method")
	}
	if n.Value != cm.Value {
		return validation.NewFieldError("Value", "cannot update value of contact method")
	}
	if n.UserID != cm.UserID {
		return validation.NewFieldError("UserID", "cannot update owner of contact method")
	}

	if permission.Admin(ctx) {
		err = gadb.New(dbtx).UpdateContactMethod(ctx, gadb.UpdateContactMethodParams{ID: uuid.MustParse(n.ID), Name: n.Name, Disabled: n.Disabled, EnableStatusUpdates: n.StatusUpdates})
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.MatchUser(cm.UserID))
	if err != nil {
		return err
	}

	err = gadb.New(dbtx).UpdateContactMethod(ctx, gadb.UpdateContactMethodParams{ID: uuid.MustParse(n.ID), Name: n.Name, Disabled: n.Disabled, EnableStatusUpdates: n.StatusUpdates})
	return err
}

// FindMany will fetch all contact methods matching the given ids.
func (s *Store) FindMany(ctx context.Context, ids []string) ([]ContactMethod, error) {
	uids, err := validate.ParseManyUUID("ContactMethodID", ids, 50)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(s.db).FindManyContactMethod(ctx, uids)
	if err != nil {
		return nil, err
	}

	cms := make([]ContactMethod, len(rows))
	for i, row := range rows {
		cms[i] = ContactMethod{
			ID:               row.ID.String(),
			Name:             row.Name,
			Type:             Type(row.Type),
			Value:            row.Value,
			Disabled:         row.Disabled,
			UserID:           row.UserID.String(),
			Pending:          row.Pending,
			StatusUpdates:    row.EnableStatusUpdates,
			lastTestVerifyAt: row.LastTestVerifyAt,
		}
	}

	return cms, nil
}

// FindAll finds all contact methods from the database associated with the given user ID.
func (s *Store) FindAll(ctx context.Context, userID string) ([]ContactMethod, error) {
	uid, err := validate.ParseUUID("ContactMethodID", userID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(s.db).FindAllContactMethod(ctx, uid)
	if err != nil {
		return nil, err
	}

	cms := make([]ContactMethod, len(rows))
	for i, row := range rows {
		cms[i] = ContactMethod{
			ID:               row.ID.String(),
			Name:             row.Name,
			Type:             Type(row.Type),
			Value:            row.Value,
			Disabled:         row.Disabled,
			UserID:           row.UserID.String(),
			Pending:          row.Pending,
			StatusUpdates:    row.EnableStatusUpdates,
			lastTestVerifyAt: row.LastTestVerifyAt,
		}
	}

	return cms, nil
}
