package graphqlapp

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
)

func (q *Query) Config(ctx context.Context, all *bool) ([]graphql2.ConfigValue, error) {
	perm := []permission.Checker{permission.System, permission.Admin}
	var publicOnly bool
	if all == nil || !*all {
		publicOnly = true
		perm = append(perm, permission.User)
	}

	err := permission.LimitCheckAny(ctx, perm...)
	if err != nil {
		return nil, err
	}

	if publicOnly {
		return graphql2.MapPublicConfigValues(q.ConfigStore.Config()), nil
	}

	return graphql2.MapConfigValues(q.ConfigStore.Config()), nil
}

func (m *Mutation) SetConfig(ctx context.Context, input []graphql2.ConfigValueInput) (bool, error) {
	err := m.ConfigStore.UpdateConfig(ctx, func(cfg config.Config) (config.Config, error) {
		return graphql2.ApplyConfigValues(cfg, input)
	})
	return err == nil, err
}
