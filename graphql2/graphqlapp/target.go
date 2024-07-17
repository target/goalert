package graphqlapp

import (
	context "context"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
)

type Target App

func (a *App) Target() graphql2.TargetResolver { return (*Target)(a) }

func (t *Target) Name(ctx context.Context, raw *assignment.RawTarget) (string, error) {
	if raw.Name != "" {
		return raw.Name, nil
	}
	switch raw.Type {
	case assignment.TargetTypeEscalationPolicy:
		ep, err := (*App)(t).FindOnePolicy(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return ep.Name, nil
	case assignment.TargetTypeService:
		svc, err := (*App)(t).FindOneService(ctx, raw.ID)
		if err != nil {
			return "", err
		}
		return svc.Name, nil
	}

	dest, err := CompatTargetToDest(raw)
	if err != nil {
		return "", err
	}

	info, err := t.DestReg.DisplayInfo(ctx, dest)
	if err != nil {
		return "", err
	}

	return info.Text, nil
}
