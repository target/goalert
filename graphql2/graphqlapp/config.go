package graphqlapp

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
)

func (a *Query) IntegrationKeyTypes(ctx context.Context) ([]graphql2.IntegrationKeyTypeInfo, error) {
	cfg := config.FromContext(ctx)
	return []graphql2.IntegrationKeyTypeInfo{
		{ID: "email", Name: "Email", Label: "Email Address", Enabled: cfg.EmailIngressEnabled()},
		{ID: "generic", Name: "Generic API", Label: "Generic Webhook URL", Enabled: true},
		{ID: "grafana", Name: "Grafana", Label: "Grafana Webhook URL", Enabled: true},
		{ID: "site24x7", Name: "Site 24x7", Label: "Site24x7 Webhook URL", Enabled: true},
		{ID: "prometheusAlertmanager", Label: "Alertmanager Webhook URL", Name: "Prometheus Alertmanager", Enabled: true},
	}, nil
}

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
func (q *Query) ConfigHints(ctx context.Context) ([]graphql2.ConfigHint, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}

	return graphql2.MapConfigHints(q.ConfigStore.Config().Hints()), nil
}

func (m *Mutation) SetConfig(ctx context.Context, input []graphql2.ConfigValueInput) (bool, error) {
	err := m.ConfigStore.UpdateConfig(ctx, func(cfg config.Config) (config.Config, error) {
		return graphql2.ApplyConfigValues(cfg, input)
	})
	return err == nil, err
}
