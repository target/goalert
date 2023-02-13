package graphqlapp

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type ContactMethod App

func (a *App) UserContactMethod() graphql2.UserContactMethodResolver {
	return (*ContactMethod)(a)
}

func (a *ContactMethod) Value(ctx context.Context, obj *contactmethod.ContactMethod) (string, error) {
	if obj.Type != contactmethod.TypeWebhook {
		return obj.Value, nil
	}

	u, err := url.Parse(obj.Value)
	if err != nil {
		return "", err
	}
	return webhook.MaskURLPass(u), nil
}

func (a *ContactMethod) FormattedValue(ctx context.Context, obj *contactmethod.ContactMethod) (string, error) {
	return a.FormatDestFunc(ctx, notification.ScannableDestType{CM: obj.Type}.DestType(), obj.Value), nil
}

func (a *ContactMethod) LastTestMessageState(ctx context.Context, obj *contactmethod.ContactMethod) (*graphql2.NotificationState, error) {
	t := obj.LastTestVerifyAt()
	if t.IsZero() {
		return nil, nil
	}

	status, _, err := a.NotificationStore.LastMessageStatus(ctx, notification.MessageTypeTest, obj.ID, t)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	return notificationStateFromSendResult(status.Status, a.FormatDestFunc(ctx, status.DestType, status.SrcValue)), nil
}

func (a *ContactMethod) LastVerifyMessageState(ctx context.Context, obj *contactmethod.ContactMethod) (*graphql2.NotificationState, error) {
	t := obj.LastTestVerifyAt()
	if t.IsZero() {
		return nil, nil
	}

	status, _, err := a.NotificationStore.LastMessageStatus(ctx, notification.MessageTypeVerification, obj.ID, t)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	return notificationStateFromSendResult(status.Status, a.FormatDestFunc(ctx, status.DestType, status.SrcValue)), nil
}

func (q *Query) UserContactMethod(ctx context.Context, id string) (*contactmethod.ContactMethod, error) {
	return (*App)(q).FindOneCM(ctx, id)
}

func (m *Mutation) CreateUserContactMethod(ctx context.Context, input graphql2.CreateUserContactMethodInput) (*contactmethod.ContactMethod, error) {
	var cm *contactmethod.ContactMethod
	cfg := config.FromContext(ctx)

	if input.Type == contactmethod.TypeWebhook && !cfg.ValidWebhookURL(input.Value) {
		return nil, validation.NewFieldError("value", "URL not allowed by administrator")
	}

	if input.Type == contactmethod.TypeSlackDM {
		if !expflag.ContextHas(ctx, expflag.SlackDM) {
			return nil, validation.NewFieldError("type", "Slack DMs are not enabled")
		}
		if strings.HasPrefix(input.Value, "@") {
			return nil, validation.NewFieldError("value", "Use 'Copy member ID' from your Slack profile to get your user ID.")
		}
		formatted := m.FormatDestFunc(ctx, notification.DestTypeSlackDM, input.Value)
		if !strings.HasPrefix(formatted, "@") {
			return nil, validation.NewFieldError("value", "Not a valid Slack user ID")
		}
	}

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
		if errors.Is(err, sql.ErrNoRows) {
			return validation.NewFieldError("id", "contact method not found")
		}
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
