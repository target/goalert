package graphqlapp

import (
	context "context"
	"database/sql"
	"github.com/target/goalert/validation/validate"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/validation"
)

func (q *Query) UserContactMethod(ctx context.Context, id string) (*contactmethod.ContactMethod, error) {
	return (*App)(q).FindOneCM(ctx, id)
}

func (m *Mutation) CreateUserContactMethod(ctx context.Context, input graphql2.CreateUserContactMethodInput) (*contactmethod.ContactMethod, error) {
	var cm *contactmethod.ContactMethod
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		cm, err = m.CMStore.CreateTx(ctx, tx, &contactmethod.ContactMethod{
			Name:     input.Name,
			Type:     input.Type,
			UserID:   input.UserID,
			Value:    input.Value,
			Disabled: true,
		})
		if err != nil {
			return err
		}

		if input.NewUserNotificationRule != nil {
			input.NewUserNotificationRule.UserID = &input.UserID
			input.NewUserNotificationRule.ContactMethodID = &cm.ID

			_, err = m.CreateUserNotificationRule(ctx, *input.NewUserNotificationRule)

			if err != nil {
				return validation.AddPrefix("newUserNotificationRule.", err)
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (m *Mutation) UpdateUserContactMethod(ctx context.Context, input graphql2.UpdateUserContactMethodInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cm, err := m.CMStore.FindOneTx(ctx, tx, input.ID)
		if err != nil {
			return err
		}
		if input.Name != nil {
			cm.Name = *input.Name
		}
		if input.Value != nil {
			cm.Value = *input.Value
		}
		return m.CMStore.UpdateTx(ctx, tx, cm)
	})
	return err == nil, err
}

func (m *Mutation) SendContactMethodVerification(ctx context.Context, input graphql2.SendContactMethodVerificationInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		return m.NotificationStore.SendContactMethodVerification(ctx, input.ContactMethodID)
	})
	return err == nil, err
}

func (m *Mutation) VerifyContactMethod(ctx context.Context, input graphql2.VerifyContactMethodInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		err := validate.Range("Code", input.Code, 100000, 999999)
		if err != nil {
			// return "must be 6 digits" error as we care about # of digits, not the code's actual value
			return validation.NewFieldError("Code", "must be 6 digits")
		}

		return m.NotificationStore.VerifyContactMethod(ctx, input.ContactMethodID, input.Code)
	})
	return err == nil, err
}
