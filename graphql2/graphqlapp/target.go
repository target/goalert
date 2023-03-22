package graphqlapp

import (
	context "context"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"

	"github.com/pkg/errors"
)

type Target App

func (a *App) Target() graphql2.TargetResolver { return (*Target)(a) }

func (t *Target) Name(ctx context.Context, raw *assignment.RawTarget) (string, error) {
	if raw.Name != "" {
		return raw.Name, nil
	}
	switch raw.Type {
	case assignment.TargetTypeRotation:
		r, err := (*App)(t).FindOneRotation(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return r.Name, nil
	case assignment.TargetTypeUser:
		u, err := (*App)(t).FindOneUser(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return u.Name, nil
	case assignment.TargetTypeEscalationPolicy:
		ep, err := (*App)(t).FindOnePolicy(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return ep.Name, nil
	case assignment.TargetTypeSchedule:
		sched, err := (*App)(t).FindOneSchedule(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return sched.Name, nil
	case assignment.TargetTypeService:
		svc, err := (*App)(t).FindOneService(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return svc.Name, nil

	}

	return "", errors.New("unhandled target type " + raw.Type.String())
}
