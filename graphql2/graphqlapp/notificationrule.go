package graphqlapp

import (
	context "context"
	"database/sql"
	"fmt"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/validation/validate"
)

type UserNotificationRule App

func (a *App) UserNotificationRule() graphql2.UserNotificationRuleResolver {
	return (*UserNotificationRule)(a)
}

func (m *Mutation) CreateUserNotificationRule(ctx context.Context, input graphql2.CreateUserNotificationRuleInput) (*notificationrule.NotificationRule, error) {
	nr := &notificationrule.NotificationRule{
		DelayMinutes: input.DelayMinutes,
	}

	if input.UserID != nil {
		nr.UserID = *input.UserID
	}

	if input.ContactMethodID != nil {
		id, err := validate.ParseUUID("ContactMethodID", *input.ContactMethodID)
		if err != nil {
			return nil, err
		}
		nr.ContactMethodID = id
	}

	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		nr, err = m.NRStore.CreateTx(ctx, tx, nr)
		return err
	})
	if err != nil {
		return nil, err
	}

	return nr, nil
}

func (nr *UserNotificationRule) ContactMethod(ctx context.Context, raw *notificationrule.NotificationRule) (*contactmethod.ContactMethod, error) {
	fmt.Println("raw.ContactMethodID", raw.ContactMethodID)
	return (*App)(nr).FindOneCM(ctx, raw.ContactMethodID)
}
