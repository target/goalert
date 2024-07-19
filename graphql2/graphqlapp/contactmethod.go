package graphqlapp

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type ContactMethod App

func (a *App) UserContactMethod() graphql2.UserContactMethodResolver {
	return (*ContactMethod)(a)
}

func (a *ContactMethod) Dest(ctx context.Context, obj *contactmethod.ContactMethod) (*gadb.DestV1, error) {
	switch obj.Type {
	case contactmethod.TypeSMS:
		return &gadb.DestV1{
			Type: twilio.DestTypeTwilioSMS,
			Args: map[string]string{twilio.FieldPhoneNumber: obj.Value},
		}, nil
	case contactmethod.TypeVoice:
		return &gadb.DestV1{
			Type: twilio.DestTypeTwilioVoice,
			Args: map[string]string{twilio.FieldPhoneNumber: obj.Value},
		}, nil
	case contactmethod.TypeEmail:
		return &gadb.DestV1{
			Type: destSMTP,
			Args: map[string]string{fieldEmailAddress: obj.Value},
		}, nil
	case contactmethod.TypeWebhook:
		return &gadb.DestV1{
			Type: webhook.DestTypeWebhook,
			Args: map[string]string{webhook.FieldWebhookURL: obj.Value},
		}, nil
	case contactmethod.TypeSlackDM:
		return &gadb.DestV1{
			Type: slack.DestTypeSlackDirectMessage,
			Args: map[string]string{slack.FieldSlackUserID: obj.Value},
		}, nil
	}

	return nil, validation.NewGenericError("unsupported data type")
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

func (a *ContactMethod) StatusUpdates(ctx context.Context, obj *contactmethod.ContactMethod) (graphql2.StatusUpdateState, error) {
	if obj.Type.StatusUpdatesAlways() {
		return graphql2.StatusUpdateStateEnabledForced, nil
	}

	if obj.Type.StatusUpdatesNever() {
		return graphql2.StatusUpdateStateDisabledForced, nil
	}

	if obj.StatusUpdates {
		return graphql2.StatusUpdateStateEnabled, nil
	}

	return graphql2.StatusUpdateStateDisabled, nil
}

func (a *ContactMethod) FormattedValue(ctx context.Context, obj *contactmethod.ContactMethod) (string, error) {
	return a.FormatDestFunc(ctx, notification.ScannableDestType{CM: obj.Type}.DestType(), obj.Value), nil
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

	return notificationStateFromSendResult(status.Status, a.FormatDestFunc(ctx, status.DestType, status.SrcValue)), nil
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

	return notificationStateFromSendResult(status.Status, a.FormatDestFunc(ctx, status.DestType, status.SrcValue)), nil
}

func (q *Query) UserContactMethod(ctx context.Context, idStr string) (*contactmethod.ContactMethod, error) {
	id, err := validate.ParseUUID("ID", idStr)
	if err != nil {
		return nil, err
	}
	return (*App)(q).FindOneCM(ctx, id)
}

func (m *Mutation) CreateUserContactMethod(ctx context.Context, input graphql2.CreateUserContactMethodInput) (*contactmethod.ContactMethod, error) {
	var cm *contactmethod.ContactMethod
	cfg := config.FromContext(ctx)

	if input.Dest != nil {
		err := validate.IDName("input.name", input.Name)
		if err != nil {
			addInputError(ctx, err)
			return nil, errAlreadySet
		}
		if err := (*App)(m).ValidateDestination(ctx, "input.dest", input.Dest); err != nil {
			return nil, err
		}
		t, v := CompatDestToCMTypeVal(*input.Dest)
		input.Type = &t
		input.Value = &v
	}

	if input.Type == nil || input.Value == nil {
		return nil, validation.NewFieldError("dest", "must be provided (or type and value)")
	}

	if *input.Type == contactmethod.TypeWebhook && !cfg.ValidWebhookURL(*input.Value) {
		return nil, validation.NewFieldError("value", "URL not allowed by administrator")
	}

	if *input.Type == contactmethod.TypeSlackDM {
		if strings.HasPrefix(*input.Value, "@") {
			return nil, validation.NewFieldError("value", "Use 'Copy member ID' from your Slack profile to get your user ID.")
		}
		formatted := m.FormatDestFunc(ctx, notification.DestTypeSlackDM, *input.Value)
		if !strings.HasPrefix(formatted, "@") {
			return nil, validation.NewFieldError("value", "Not a valid Slack user ID")
		}
	}

	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		cm, err = m.CMStore.Create(ctx, tx, &contactmethod.ContactMethod{
			Name:     input.Name,
			Type:     *input.Type,
			UserID:   input.UserID,
			Value:    *input.Value,
			Disabled: true,

			StatusUpdates: input.EnableStatusUpdates != nil && *input.EnableStatusUpdates,
		})
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
		if input.Value != nil {
			cm.Value = *input.Value
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
