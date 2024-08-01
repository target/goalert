package graphqlapp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/email"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type (
	ContactMethod App
)

func (a *App) UserContactMethod() graphql2.UserContactMethodResolver {
	return (*ContactMethod)(a)
}

func (a *ContactMethod) Type(ctx context.Context, obj *contactmethod.ContactMethod) (*graphql2.ContactMethodType, error) {
	cmType, _ := CompatDestToCMTypeVal(obj.Dest)
	return &cmType, nil
}

func (a *ContactMethod) Value(ctx context.Context, obj *contactmethod.ContactMethod) (string, error) {
	_, cmVal := CompatDestToCMTypeVal(obj.Dest)
	return cmVal, nil
}

func (a *ContactMethod) StatusUpdates(ctx context.Context, obj *contactmethod.ContactMethod) (graphql2.StatusUpdateState, error) {
	info, err := a.DestReg.TypeInfo(ctx, obj.Dest.Type)
	if err != nil {
		return "", err
	}

	if !info.SupportsStatusUpdates {
		return graphql2.StatusUpdateStateDisabledForced, nil
	}

	if info.StatusUpdatesRequired {
		return graphql2.StatusUpdateStateEnabledForced, nil
	}

	if obj.StatusUpdates {
		return graphql2.StatusUpdateStateEnabled, nil
	}

	return graphql2.StatusUpdateStateDisabled, nil
}

func (a *ContactMethod) FormattedValue(ctx context.Context, obj *contactmethod.ContactMethod) (string, error) {
	info, err := a.DestReg.DisplayInfo(ctx, obj.Dest)
	if err != nil {
		return "", err
	}
	return info.Text, nil
}

func (a *ContactMethod) LastTestMessageState(ctx context.Context, obj *contactmethod.ContactMethod) (*graphql2.NotificationState, error) {
	t := obj.LastTestVerifyAt()
	if t.IsZero() {
		return nil, nil
	}

	status, _, err := a.NotificationStore.LastMessageStatus(ctx, notification.MessageTypeTest, obj.ID.String(), t)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	return notificationStateFromSendResult(status.Status, status.SrcValue), nil
}

func (a *ContactMethod) LastVerifyMessageState(ctx context.Context, obj *contactmethod.ContactMethod) (*graphql2.NotificationState, error) {
	t := obj.LastTestVerifyAt()
	if t.IsZero() {
		return nil, nil
	}

	status, _, err := a.NotificationStore.LastMessageStatus(ctx, notification.MessageTypeVerification, obj.ID.String(), t)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	return notificationStateFromSendResult(status.Status, status.SrcValue), nil
}

func (q *Query) UserContactMethod(ctx context.Context, idStr string) (*contactmethod.ContactMethod, error) {
	id, err := validate.ParseUUID("ID", idStr)
	if err != nil {
		return nil, err
	}
	return (*App)(q).FindOneCM(ctx, id)
}

func (m *Mutation) CreateUserContactMethod(ctx context.Context, input graphql2.CreateUserContactMethodInput) (*contactmethod.ContactMethod, error) {
	cm := &contactmethod.ContactMethod{
		Name:          input.Name,
		UserID:        input.UserID,
		Disabled:      true,
		StatusUpdates: input.EnableStatusUpdates != nil && *input.EnableStatusUpdates,
	}

	if input.Dest != nil {
		if err := (*App)(m).ValidateDestination(ctx, "input.dest", input.Dest); err != nil {
			return nil, err
		}
		cm.Dest = *input.Dest
	} else if input.Type != nil && input.Value != nil {
		switch *input.Type {
		case graphql2.ContactMethodTypeEmail:
			cm.Dest = email.NewEmailDest(*input.Value)
		case graphql2.ContactMethodTypeSms:
			cm.Dest = twilio.NewSMSDest(*input.Value)
		case graphql2.ContactMethodTypeVoice:
			cm.Dest = twilio.NewVoiceDest(*input.Value)
		case graphql2.ContactMethodTypeSLACkDm:
			cm.Dest = slack.NewDirectMessageDest(*input.Value)
		case graphql2.ContactMethodTypeWebhook:
			cm.Dest = webhook.NewWebhookDest(*input.Value)
		}

		return nil, validation.NewFieldError("input.Type", "unsupported type")
	} else {
		return nil, validation.NewFieldError("input", "must provide either dest or type/value")
	}

	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		cm, err = m.CMStore.Create(ctx, tx, cm)
		if err != nil {
			return err
		}

		if input.NewUserNotificationRule != nil {
			input.NewUserNotificationRule.UserID = &input.UserID
			str := cm.ID.String()
			input.NewUserNotificationRule.ContactMethodID = &str

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
	if input.Value != nil {
		return false, validation.NewFieldError("input.value", "cannot update value")
	}

	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		id, err := validate.ParseUUID("ID", input.ID)
		if err != nil {
			return err
		}
		cm, err := m.CMStore.FindOne(ctx, tx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return validation.NewFieldError("id", "contact method not found")
		}
		if err != nil {
			return err
		}
		if input.Name != nil {
			err := validate.IDName("input.name", *input.Name)
			if err != nil {
				addInputError(ctx, err)
				return errAlreadySet
			}
			cm.Name = *input.Name
		}

		if input.EnableStatusUpdates != nil {
			cm.StatusUpdates = *input.EnableStatusUpdates
		}

		return m.CMStore.Update(ctx, tx, cm)
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
