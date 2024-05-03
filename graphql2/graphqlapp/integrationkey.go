package graphqlapp

import (
	context "context"
	"database/sql"
	"net/url"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation/validate"
)

type IntegrationKey App

func (a *App) IntegrationKey() graphql2.IntegrationKeyResolver { return (*IntegrationKey)(a) }

func (q *Query) IntegrationKey(ctx context.Context, id string) (*integrationkey.IntegrationKey, error) {
	return q.IntKeyStore.FindOne(ctx, id)
}

func (m *Mutation) UpdateKeyConfig(ctx context.Context, input graphql2.UpdateKeyConfigInput) (bool, error) {
	return false, nil
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
		if input.ExternalSystemName != nil {
			key.ExternalSystemName = *input.ExternalSystemName
		}
		key, err = m.IntKeyStore.Create(ctx, tx, key)
		return err
	})
	return key, err
}

func (key *IntegrationKey) Config(ctx context.Context, raw *integrationkey.IntegrationKey) (*graphql2.KeyConfig, error) {
	id, err := validate.ParseUUID("IntegrationKey.ID", raw.ID)
	if err != nil {
		return nil, err
	}

	cfg, err := key.IntKeyStore.Config(ctx, key.DB, id)
	if err != nil {
		return nil, err
	}

	var rules []graphql2.KeyRule
	for _, r := range cfg.Rules {
		var actions []graphql2.Action
		for _, a := range r.Actions {
			actions = append(actions, graphql2.Action{
				Dest:   &graphql2.Destination{Type: a.Type, Values: mapToFieldValue(a.StaticParams)},
				Params: mapToParams(a.DynamicParams),
			})
		}
		rules = append(rules, graphql2.KeyRule{
			ID:   r.ID.String(),
			Name: r.Name,
			// Description:   r.Description,
			ConditionExpr: r.ConditionExpr,
			Dedup: &graphql2.DedupConfig{
				DedupExpr:   r.DedupConfig.IDExpr,
				DedupWindow: timeutil.ISODuration{SecondPart: float64(r.DedupConfig.WindowSeconds)},
			},
			Action: actions,
		})
	}

	var supp []graphql2.SuppressionWindow
	n := time.Now()
	for _, s := range cfg.Suppression {
		supp = append(supp, graphql2.SuppressionWindow{
			Start:      s.Start,
			End:        s.End,
			Active:     !s.Start.After(n) && s.End.Before(n),
			FilterExpr: s.FilterExpr,
		})
	}

	var actions []graphql2.Action
	for _, a := range cfg.DefaultActions {
		actions = append(actions, graphql2.Action{
			Dest:   &graphql2.Destination{Type: a.Type, Values: mapToFieldValue(a.StaticParams)},
			Params: mapToParams(a.DynamicParams),
		})
	}

	return &graphql2.KeyConfig{
		StopAtFirstRule:    cfg.StopOnFirstRule,
		Rules:              rules,
		SuppressionWindows: supp,
		DefaultActions:     actions,
	}, nil
}

func mapToFieldValue(m map[string]string) []graphql2.FieldValuePair {
	res := make([]graphql2.FieldValuePair, 0, len(m))
	for k, v := range m {
		res = append(res, graphql2.FieldValuePair{
			FieldID: k,
			Value:   v,
		})
	}
	return res
}

func mapToParams(m map[string]string) []graphql2.DynamicParam {
	res := make([]graphql2.DynamicParam, 0, len(m))
	for k, v := range m {
		res = append(res, graphql2.DynamicParam{
			ParamID: k,
			Expr:    v,
		})
	}
	return res
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
