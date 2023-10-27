package contactmethod

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Store implements the lookup and management of ContactMethods against a *sql.Store backend.
type Store struct {
}

func (s *Store) MetadataByTypeValue(ctx context.Context, dbtx gadb.DBTX, typ Type, value string) (*Metadata, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	data, err := gadb.New(dbtx).ContactMethodMetaTV(ctx, gadb.ContactMethodMetaTVParams{Type: gadb.EnumUserContactMethodType(typ), Value: value})
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

func (s *Store) SetCarrierV1MetadataByTypeValue(ctx context.Context, dbtx gadb.DBTX, typ Type, value string, newM *Metadata) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}

	data, err := json.Marshal(newM.CarrierV1)
	if err != nil {
		return err
	}

	err = gadb.New(dbtx).ContactMethodUpdateMetaTV(ctx, gadb.ContactMethodUpdateMetaTVParams{Type: gadb.EnumUserContactMethodType(typ), Value: value, CarrierV1: data})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) EnableByValue(ctx context.Context, dbtx gadb.DBTX, t Type, v string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	c := ContactMethod{Name: "Enable", Type: t, Value: v}
	n, err := c.Normalize()
	if err != nil {
		return err
	}

	id, err := gadb.New(dbtx).ContactMethodEnable(ctx, gadb.ContactMethodEnableParams{Type: gadb.EnumUserContactMethodType(n.Type), Value: n.Value})

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method START code received.")
	}

	return err
}

func (s *Store) DisableByValue(ctx context.Context, dbtx gadb.DBTX, t Type, v string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	c := ContactMethod{Name: "Disable", Type: t, Value: v}
	n, err := c.Normalize()
	if err != nil {
		return err
	}

	id, err := gadb.New(dbtx).ContactMethodDisable(ctx, gadb.ContactMethodDisableParams{Type: gadb.EnumUserContactMethodType(n.Type), Value: v})

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

	err = gadb.New(dbtx).ContactMethodAdd(ctx, gadb.ContactMethodAddParams{
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

	rows, err := gadb.New(dbtx).ContactMethodLookupUserID(ctx, uids)
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

	row, err := gadb.New(dbtx).ContactMethodFindOneUpdate(ctx, methodUUID)
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
		err = gadb.New(dbtx).ContactMethodUpdate(ctx, gadb.ContactMethodUpdateParams{ID: uuid.MustParse(n.ID), Name: n.Name, Disabled: n.Disabled, EnableStatusUpdates: n.StatusUpdates})
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.MatchUser(cm.UserID))
	if err != nil {
		return err
	}

	err = gadb.New(dbtx).ContactMethodUpdate(ctx, gadb.ContactMethodUpdateParams{ID: uuid.MustParse(n.ID), Name: n.Name, Disabled: n.Disabled, EnableStatusUpdates: n.StatusUpdates})

	return err
}

// FindMany will fetch all contact methods matching the given ids.
func (s *Store) FindMany(ctx context.Context, dbtx gadb.DBTX, ids []string) ([]ContactMethod, error) {
	uids, err := validate.ParseManyUUID("ContactMethodID", ids, 50)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(dbtx).ContactMethodFindMany(ctx, uids)
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
func (s *Store) FindAll(ctx context.Context, dbtx gadb.DBTX, userID string) ([]ContactMethod, error) {
	uid, err := validate.ParseUUID("ContactMethodID", userID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(dbtx).ContactMethodFindAll(ctx, uid)
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
