package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"github.com/ttacon/libphonenumber"
)

type ContactMethod App

func (a *App) UserContactMethod() graphql2.UserContactMethodResolver {
	return (*ContactMethod)(a)
}

func (a *ContactMethod) FormattedValue(ctx context.Context, obj *contactmethod.ContactMethod) (string, error) {
	formatted := obj.Value
	switch obj.Type {
	case contactmethod.TypeSMS, contactmethod.TypeVoice:
		num, err := libphonenumber.Parse(obj.Value, "")
		if err != nil {
			log.Log(ctx, err)
			break
		}
		formatted = libphonenumber.Format(num, libphonenumber.INTERNATIONAL)
	}
	return formatted, nil
}

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
	err := m.NotificationStore.SendContactMethodVerification(ctx, input.ContactMethodID)
	return err == nil, err
}

func (m *Mutation) VerifyContactMethod(ctx context.Context, input graphql2.VerifyContactMethodInput) (bool, error) {
	err := validate.Range("Code", input.Code, 100000, 999999)
	if err != nil {
		// return "must be 6 digits" error as we care about # of digits, not the code's actual value
		return false, validation.NewFieldError("Code", "must be 6 digits")
	}

	err = m.NotificationStore.VerifyContactMethod(ctx, input.ContactMethodID, input.Code)
	return err == nil, err
}

func (a *Query) SendTestStatus(ctx context.Context, cmID string) (*graphql2.NotificationState, error) {
	if cmID == "" {
		return nil, validation.NewFieldError("cmID", "field is required")
	}

	s, err := a.CMStore.FindLastStatus(ctx, cmID)
	if err != nil {
		return nil, err
	}

	return notificationStateFromStatus(*s)
}
