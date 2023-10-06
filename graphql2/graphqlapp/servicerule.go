package graphqlapp

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/service/rule"
)

type ServiceRule App

func (a *App) ServiceRule() graphql2.ServiceRuleResolver { return (*ServiceRule)(a) }

func (q *Query) ServiceRule(ctx context.Context, id string) (*rule.Rule, error) {
	return (*App)(q).FindOneServiceRule(ctx, id)
}

func (s *ServiceRule) IntegrationKeys(ctx context.Context, r *rule.Rule) ([]integrationkey.IntegrationKey, error) {
	return s.IntKeyStore.FindAllByServiceRule(ctx, r.ID)
}

func (m *Mutation) CreateServiceRule(ctx context.Context, input graphql2.CreateServiceRuleInput) (r *rule.Rule, err error) {
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		actions := []map[string]interface{}{}
		err = json.Unmarshal([]byte(input.Actions), &actions)
		if err != nil {
			return err
		}
		r, err = m.ServiceRuleStore.Create(ctx, tx, rule.Rule{
			Name:            input.Name,
			ServiceID:       input.ServiceID,
			FilterString:    input.Filter,
			SendAlert:       input.SendAlert,
			Actions:         actions,
			IntegrationKeys: input.IntegrationKeys,
		})
		return err
	})
	return
}

func (m *Mutation) UpdateServiceRule(ctx context.Context, input graphql2.UpdateServiceRuleInput) (r *rule.Rule, err error) {
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		actions := []map[string]interface{}{}
		err = json.Unmarshal([]byte(input.Actions), &actions)
		if err != nil {
			return err
		}
		r, err = m.ServiceRuleStore.Update(ctx, tx, rule.Rule{
			ID:              input.ID,
			Name:            input.Name,
			FilterString:    input.Filter,
			SendAlert:       input.SendAlert,
			Actions:         actions,
			IntegrationKeys: input.IntegrationKeys,
		})
		return err
	})
	return
}

func (m *Mutation) DeleteServiceRule(ctx context.Context, id string) (success bool, err error) {
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		err = m.ServiceRuleStore.Delete(ctx, tx, id)
		return err
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
