package graphqlapp

import (
	context "context"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"

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
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

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
			err = errors.Wrap(a.IntKeyStore.DeleteManyTx(ctx, tx, ids), "delete integration keys")
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
			return false, validation.NewFieldError("type", "unsupported type "+typ.String())
		}
		if err != nil {
			return false, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}
