package graphqlapp

import (
	context "context"
	"database/sql"
	"net/url"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/search"
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
	case integrationkey.TypeSite24x7:
		return cfg.CallbackURL("/api/v2/site24x7/incoming", q), nil
	case integrationkey.TypePrometheusAlertmanager:
		return cfg.CallbackURL("/api/v2/prometheusalertmanager/incoming", q), nil
	case integrationkey.TypeEmail:
		if !cfg.EmailIngressEnabled() {
			return "", nil
		}
		return "mailto:" + raw.ID + "@" + cfg.EmailIngressDomain(), nil
	}

	return "", nil
}

func (q *Query) IntegrationKeys(ctx context.Context, input *graphql2.IntegrationKeySearchOptions) (conn *graphql2.IntegrationKeyConnection, err error) {
	if input == nil {
		input = &graphql2.IntegrationKeySearchOptions{}
	}

	var opts integrationkey.InKeySearchOptions
	if input.Search != nil {
		opts.Search = *input.Search
	}
	opts.Omit = input.Omit
	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &opts)
		if err != nil {
			return conn, err
		}
	}
	if input.First != nil {
		opts.Limit = *input.First
	}
	if opts.Limit == 0 {
		opts.Limit = 15
	}

	opts.Limit++
	intKeys, err := q.IntKeyStore.Search(ctx, &opts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.IntegrationKeyConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(intKeys) == opts.Limit {
		intKeys = intKeys[:len(intKeys)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(intKeys) > 0 {
		lastKey := intKeys[len(intKeys)-1]
		opts.After = lastKey.Name

		cur, err := search.Cursor(opts)
		if err != nil {
			return nil, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = intKeys
	return conn, err
}
