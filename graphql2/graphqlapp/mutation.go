package graphqlapp

import (
	context "context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type Mutation App

func (a *App) Mutation() graphql2.MutationResolver { return (*Mutation)(a) }

func (a *Mutation) SetFavorite(ctx context.Context, input graphql2.SetFavoriteInput) (bool, error) {
	var err error
	if input.Favorite {
		err = a.FavoriteStore.Set(ctx, permission.UserID(ctx), input.Target)
	} else {
		err = a.FavoriteStore.Unset(ctx, permission.UserID(ctx), input.Target)
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Mutation) LinkAccount(ctx context.Context, token string) (bool, error) {
	err := a.AuthLinkStore.LinkAccount(ctx, token)
	return err == nil, err
}

func (a *Mutation) SetScheduleOnCallNotificationRules(ctx context.Context, input graphql2.SetScheduleOnCallNotificationRulesInput) (bool, error) {
	schedID, err := parseUUID("ScheduleID", input.ScheduleID)
	if err != nil {
		return false, err
	}

	err = withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		rules := make([]schedule.OnCallNotificationRule, 0, len(input.Rules))
		for i, r := range input.Rules {
			err := validate.OneOf(fmt.Sprintf("Rules[%d].Target.Type", i), r.Target.Type, assignment.TargetTypeSlackChannel, assignment.TargetTypeSlackUserGroup, assignment.TargetTypeChanWebhook)
			if err != nil {
				return err
			}

			var nfyChan *notificationchannel.Channel
			switch r.Target.Type {
			case assignment.TargetTypeSlackUserGroup:
				grpID, chanID, _ := strings.Cut(r.Target.ID, ":")
				grp, err := a.SlackStore.UserGroup(ctx, grpID)
				if err != nil {
					return validation.WrapError(err)
				}
				ch, err := a.SlackStore.Channel(ctx, chanID)
				if err != nil {
					return validation.WrapError(err)
				}

				nfyChan = &notificationchannel.Channel{
					Type:  notificationchannel.TypeSlackUG,
					Name:  fmt.Sprintf("%s (%s)", grp.Handle, ch.Name),
					Value: r.Target.ID,
				}
			case assignment.TargetTypeSlackChannel:
				ch, err := a.SlackStore.Channel(ctx, r.Target.ID)
				if err != nil {
					return err
				}

				nfyChan = &notificationchannel.Channel{
					Type:  notificationchannel.TypeSlackChan,
					Name:  ch.Name,
					Value: ch.ID,
				}
			case assignment.TargetTypeChanWebhook:
				url, err := url.Parse(r.Target.ID)
				if err != nil {
					return validation.NewFieldError("Rules[%d].Target.ID", "Invalid URL format")
				}
				url.RawQuery = ""
				if len(url.Path) > 15 {
					url.Path = url.Path[:12] + "..."
				}

				cfg := config.FromContext(ctx)
				if !cfg.ValidWebhookURL(r.Target.ID) {
					return validation.NewFieldError("Rules[%d].Target.ID", "URL not allowed by administrator")
				}

				nfyChan = &notificationchannel.Channel{
					Type:  notificationchannel.TypeWebhook,
					Name:  webhook.MaskURLPass(url),
					Value: r.Target.ID,
				}
			}

			r.ChannelID, err = a.NCStore.MapToID(ctx, tx, nfyChan)
			if err != nil {
				return err
			}
			rules = append(rules, r.OnCallNotificationRule)
		}

		return a.ScheduleStore.SetOnCallNotificationRules(ctx, tx, schedID, rules)
	})

	return err == nil, err
}

func (a *Mutation) SetTemporarySchedule(ctx context.Context, input graphql2.SetTemporaryScheduleInput) (bool, error) {
	schedID, err := parseUUID("ScheduleID", input.ScheduleID)
	if err != nil {
		return false, err
	}

	tmp := schedule.TemporarySchedule{
		Start:  input.Start,
		End:    input.End,
		Shifts: input.Shifts,
	}

	var clearSet bool
	if input.ClearStart != nil || input.ClearEnd != nil {
		if input.ClearStart == nil {
			return false, validation.NewFieldError("ClearStart", "must be set if ClearEnd is set")
		}
		if input.ClearEnd == nil {
			return false, validation.NewFieldError("ClearEnd", "must be set if ClearStart is set")
		}
		clearSet = true
	}

	err = withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		if clearSet {
			return a.ScheduleStore.SetClearTemporarySchedule(ctx, tx, schedID, tmp, *input.ClearStart, *input.ClearEnd)
		}

		return a.ScheduleStore.SetTemporarySchedule(ctx, tx, schedID, tmp)
	})

	return err == nil, err
}

func (a *Mutation) ClearTemporarySchedules(ctx context.Context, input graphql2.ClearTemporarySchedulesInput) (bool, error) {
	schedID, err := parseUUID("ScheduleID", input.ScheduleID)
	if err != nil {
		return false, err
	}

	err = withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		return a.ScheduleStore.ClearTemporarySchedules(ctx, tx, schedID, input.Start, input.End)
	})

	return err == nil, err
}

func (a *Mutation) TestContactMethod(ctx context.Context, id string) (bool, error) {
	err := a.NotificationStore.SendContactMethodTest(ctx, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *Mutation) AddAuthSubject(ctx context.Context, input user.AuthSubject) (bool, error) {
	err := a.UserStore.AddAuthSubjectTx(ctx, nil, &input)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Mutation) DeleteAuthSubject(ctx context.Context, input user.AuthSubject) (bool, error) {
	err := a.UserStore.DeleteAuthSubjectTx(ctx, nil, &input)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Mutation) EndAllAuthSessionsByCurrentUser(ctx context.Context) (bool, error) {
	err := a.AuthHandler.EndAllUserSessionsTx(ctx, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Mutation) DeleteAll(ctx context.Context, input []assignment.RawTarget) (bool, error) {
	// Retry because deleting frequently can cause a deadlock
	// under heavy load.
	err := retry.DoTemporaryError(func(int) error {
		return a.tryDeleteAll(ctx, input)
	},
		retry.Log(ctx),
		retry.Limit(5),
		retry.FibBackoff(time.Second),
	)

	return err == nil, err
}

func (a *Mutation) tryDeleteAll(ctx context.Context, input []assignment.RawTarget) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "graphql: delete all", tx)

	m := make(map[assignment.TargetType][]string)
	for _, tgt := range input {
		m[tgt.TargetType()] = append(m[tgt.TargetType()], tgt.TargetID())
	}

	order := []assignment.TargetType{
		assignment.TargetTypeRotation,
		assignment.TargetTypeUserOverride,
		assignment.TargetTypeSchedule,
		assignment.TargetTypeCalendarSubscription,
		assignment.TargetTypeUser,
		assignment.TargetTypeIntegrationKey,
		assignment.TargetTypeHeartbeatMonitor,
		assignment.TargetTypeService,
		assignment.TargetTypeEscalationPolicy,
		assignment.TargetTypeNotificationRule,
		assignment.TargetTypeContactMethod,
		assignment.TargetTypeUserSession,
	}

	for _, typ := range order {
		ids := m[typ]
		if len(ids) == 0 {
			continue
		}
		switch typ {
		case assignment.TargetTypeUserOverride:
			err = errors.Wrap(a.OverrideStore.DeleteUserOverrideTx(ctx, tx, ids...), "delete user overrides")
		case assignment.TargetTypeUser:
			err = errors.Wrap(a.UserStore.DeleteManyTx(ctx, tx, ids), "delete users")
		case assignment.TargetTypeService:
			err = errors.Wrap(a.ServiceStore.DeleteManyTx(ctx, tx, ids), "delete services")
		case assignment.TargetTypeEscalationPolicy:
			err = errors.Wrap(a.PolicyStore.DeleteManyPoliciesTx(ctx, tx, ids), "delete escalation policies")
		case assignment.TargetTypeIntegrationKey:
			err = errors.Wrap(a.IntKeyStore.DeleteMany(ctx, tx, ids), "delete integration keys")
		case assignment.TargetTypeSchedule:
			err = errors.Wrap(a.ScheduleStore.DeleteManyTx(ctx, tx, ids), "delete schedules")
		case assignment.TargetTypeCalendarSubscription:
			err = errors.Wrap(a.CalSubStore.DeleteTx(ctx, tx, permission.UserID(ctx), ids...), "delete calendar subscriptions")
		case assignment.TargetTypeRotation:
			err = errors.Wrap(a.RotationStore.DeleteManyTx(ctx, tx, ids), "delete rotations")
		case assignment.TargetTypeContactMethod:
			err = errors.Wrap(a.CMStore.DeleteTx(ctx, tx, ids...), "delete contact methods")
		case assignment.TargetTypeNotificationRule:
			err = errors.Wrap(a.NRStore.DeleteTx(ctx, tx, ids...), "delete notification rules")
		case assignment.TargetTypeHeartbeatMonitor:
			err = errors.Wrap(a.HeartbeatStore.DeleteTx(ctx, tx, ids...), "delete heartbeat monitors")
		case assignment.TargetTypeUserSession:
			err = errors.Wrap(a.AuthHandler.EndUserSessionTx(ctx, tx, ids...), "end user sessions")
		default:
			return validation.NewFieldError("type", "unsupported type "+typ.String())
		}
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
