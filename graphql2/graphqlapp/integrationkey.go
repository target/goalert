package graphqlapp

import (
	context "context"
	"database/sql"
	"net/url"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/integrationkey"
)

type IntegrationKey App

func (a *App) IntegrationKey() graphql2.IntegrationKeyResolver { return (*IntegrationKey)(a) }

func (q *Query) IntegrationKey(ctx context.Context, id string) (*integrationkey.IntegrationKey, error) {
	return q.IntKeyStore.FindOne(ctx, id)
}
func (m *Mutation) CreateIntegrationKey(ctx context.Context, input graphql2.CreateIntegrationKeyInput) (key *integrationkey.IntegrationKey, err error) {
	var serviceID string
	if input.ServiceID != nil {
		serviceID = *input.ServiceID
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		key = &integrationkey.IntegrationKey{
			ServiceID: serviceID,
			Name:      input.Name,
			Type:      integrationkey.Type(input.Type),
		}
		key, err = m.IntKeyStore.CreateKeyTx(ctx, tx, key)
		return err
	})
	return key, err
}
func (key *IntegrationKey) Type(ctx context.Context, raw *integrationkey.IntegrationKey) (graphql2.IntegrationKeyType, error) {
	return graphql2.IntegrationKeyType(raw.Type), nil
}
func (key *IntegrationKey) Href(ctx context.Context, raw *integrationkey.IntegrationKey) (string, error) {
	cfg := config.FromContext(ctx)
	q := make(url.Values)
	q.Set("token", raw.ID)
	switch raw.Type {
	case integrationkey.TypeGeneric:
		return cfg.CallbackURL("/api/v2/generic/incoming", q), nil
	case integrationkey.TypeGrafana:
		return cfg.CallbackURL("/api/v2/grafana/incoming", q), nil
	case integrationkey.TypeEmail:
		if !cfg.Mailgun.Enable || cfg.Mailgun.EmailDomain == "" {
			return "", nil
		}
		return "mailto:" + raw.ID + "@" + cfg.Mailgun.EmailDomain, nil
	}

	return "", nil
}
