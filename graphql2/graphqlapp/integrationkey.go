package graphqlapp

import (
	context "context"
	"database/sql"
	"net/url"

	"github.com/google/uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type IntegrationKey App

type KeyConfig App

func (a *App) IntegrationKey() graphql2.IntegrationKeyResolver { return (*IntegrationKey)(a) }
func (a *App) KeyConfig() graphql2.KeyConfigResolver           { return (*KeyConfig)(a) }

func (k *KeyConfig) OneRule(ctx context.Context, key *gadb.UIKConfigV1, ruleID string) (*gadb.UIKRuleV1, error) {
	for _, r := range key.Rules {
		if r.ID.String() == ruleID {
			return &r, nil
		}
	}

	return nil, validation.NewFieldError("RuleID", "not found")
}

func (q *Query) IntegrationKey(ctx context.Context, id string) (*integrationkey.IntegrationKey, error) {
	return q.IntKeyStore.FindOne(ctx, id)
}

func (q *Query) ActionInputValidate(ctx context.Context, input gadb.UIKActionV1) (bool, error) {
	err := (*App)(q).ValidateDestination(ctx, "input.dest", &input.Dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Mutation) GenerateKeyToken(ctx context.Context, keyID string) (string, error) {
	id, err := validate.ParseUUID("ID", keyID)
	if err != nil {
		return "", err
	}
	return m.IntKeyStore.GenerateToken(ctx, m.DB, id)
}

func (m *Mutation) DeleteSecondaryToken(ctx context.Context, keyID string) (bool, error) {
	id, err := validate.ParseUUID("ID", keyID)
	if err != nil {
		return false, err
	}

	err = m.IntKeyStore.DeleteSecondaryToken(ctx, m.DB, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Mutation) PromoteSecondaryToken(ctx context.Context, keyID string) (bool, error) {
	id, err := validate.ParseUUID("ID", keyID)
	if err != nil {
		return false, err
	}

	err = m.IntKeyStore.PromoteSecondaryToken(ctx, m.DB, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (key *IntegrationKey) TokenInfo(ctx context.Context, raw *integrationkey.IntegrationKey) (*graphql2.TokenInfo, error) {
	id, err := validate.ParseUUID("ID", raw.ID)
	if err != nil {
		return nil, err
	}

	prim, sec, err := key.IntKeyStore.TokenHints(ctx, key.DB, id)
	if err != nil {
		return nil, err
	}

	return &graphql2.TokenInfo{
		PrimaryHint:   prim,
		SecondaryHint: sec,
	}, nil
}

func (m *Mutation) UpdateKeyConfig(ctx context.Context, input graphql2.UpdateKeyConfigInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		id, err := validate.ParseUUID("IntegrationKey.ID", input.KeyID)
		if err != nil {
			return err
		}

		cfg, err := m.IntKeyStore.Config(ctx, tx, id)
		if err != nil {
			return err
		}

		if input.Rules != nil {
			cfg.Rules = input.Rules
		}

		if input.SetRule != nil {
			if input.SetRule.ID == uuid.Nil {
				// Since we don't have a rule ID, we're need to create a new rule.
				cfg.Rules = append(cfg.Rules, *input.SetRule)
			} else {
				var found bool
				for i, r := range cfg.Rules {
					if r.ID == input.SetRule.ID {
						cfg.Rules[i] = *input.SetRule
						found = true
						break
					}
				}
				if !found {
					return validation.NewFieldError("SetRule.ID", "not found")
				}
			}
		}
		if input.SetRuleActions != nil {
			var found bool
			for i, r := range cfg.Rules {
				if r.ID != input.SetRuleActions.ID {
					continue
				}

				cfg.Rules[i].Actions = input.SetRuleActions.Actions
				found = true
				break
			}
			if !found {
				return validation.NewFieldError("SetRuleActions.RuleID", "not found")
			}
		}

		if input.DeleteRule != nil {
			for i, r := range cfg.Rules {
				if r.ID.String() == *input.DeleteRule {
					cfg.Rules = append(cfg.Rules[:i], cfg.Rules[i+1:]...)
					break
				}
			}
		}

		if input.DefaultActions != nil {
			cfg.DefaultActions = input.DefaultActions
		}

		err = m.IntKeyStore.SetConfig(ctx, tx, id, cfg)
		return err
	})
	if err != nil {
		return false, err
	}

	return true, nil
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

func (key *IntegrationKey) Config(ctx context.Context, raw *integrationkey.IntegrationKey) (*gadb.UIKConfigV1, error) {
	id, err := validate.ParseUUID("IntegrationKey.ID", raw.ID)
	if err != nil {
		return nil, err
	}

	return key.IntKeyStore.Config(ctx, key.DB, id)
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
	case integrationkey.TypeUniversal:
		return cfg.CallbackURL("/api/v2/uik"), nil
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
