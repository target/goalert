package rule

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sqlc-dev/pqtype"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) *Store {
	s := &Store{db: db}

	return s
}

func (s *Store) FindOne(ctx context.Context, id string) (*Rule, error) {
	ruleID, err := validate.ParseUUID("ServiceRuleID", id)
	if err != nil {
		return nil, err
	}

	row, err := gadb.New(s.db).SvcRuleFindOne(ctx, ruleID)
	if err != nil {
		return nil, errors.Wrap(err, "get rules for service")
	}

	actions := []map[string]interface{}{}
	if !row.SendAlert {
		if !row.Actions.Valid {
			return nil, fmt.Errorf("service rule has null action")
		}
		actionsRaw := row.Actions.RawMessage
		err := json.Unmarshal(actionsRaw, &actions)
		if err != nil {
			return nil, fmt.Errorf("bad actions value for service rule")
		}
	}

	return &Rule{
		ID:              row.ID.String(),
		Name:            row.Name,
		ServiceID:       row.ServiceID.String(),
		IntegrationKeys: strings.Split(row.IntegrationKeys, ","),
		FilterString:    row.Filter,
		Actions:         actions,
	}, nil
}

// FindAllByService returns all service rules associated with the given serviceID
func (s *Store) FindAllByService(ctx context.Context, serviceID string) ([]Rule, error) {
	err := permission.LimitCheckAny(ctx,
		permission.User,
		permission.MatchService(serviceID),
	)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}
	rows, err := gadb.New(s.db).SvcRuleFindManyByService(ctx, uuid.MustParse(serviceID))
	if err != nil {
		return nil, errors.Wrap(err, "find rules for service")
	}

	rules := make([]Rule, len(rows))
	for i, row := range rows {
		actions := []map[string]interface{}{}
		if !row.SendAlert {
			if !row.Actions.Valid {
				return nil, fmt.Errorf("signal rule %d has null action", i)
			}
			actionsRaw := row.Actions.RawMessage
			err := json.Unmarshal(actionsRaw, &actions)
			if err != nil {
				return nil, fmt.Errorf("bad actions value for signal rule %d", i)
			}
		}
		rules[i] = Rule{
			ID:              row.ID.String(),
			Name:            row.Name,
			ServiceID:       row.ServiceID.String(),
			IntegrationKeys: strings.Split(row.IntegrationKeys, ","),
			FilterString:    row.Filter,
			Actions:         actions,
		}
	}

	return rules, nil
}

// FindAllByIntegrationKey returns all service rules associated with the given serviceID and integrationKeyID
func (s *Store) FindAllByIntegrationKey(ctx context.Context, serviceID string, integrationKeyID string) ([]Rule, error) {
	err := permission.LimitCheckAny(ctx,
		permission.User,
		permission.MatchService(serviceID),
	)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("IntegrationKeyID", integrationKeyID)
	if err != nil {
		return nil, err
	}
	rows, err := gadb.New(s.db).SvcRuleFindManyByIntKey(ctx, gadb.SvcRuleFindManyByIntKeyParams{
		ServiceID:        uuid.MustParse(serviceID),
		IntegrationKeyID: uuid.MustParse(integrationKeyID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "get rules for integration key")
	}

	rules := make([]Rule, len(rows))
	for i, row := range rows {
		actions := []map[string]interface{}{}
		if !row.SendAlert {
			if !row.Actions.Valid {
				return nil, fmt.Errorf("signal rule %d has null action", i)
			}
			actionsRaw := row.Actions.RawMessage
			err := json.Unmarshal(actionsRaw, &actions)
			if err != nil {
				return nil, fmt.Errorf("bad actions value for signal rule %d", i)
			}
		}
		rules[i] = Rule{
			ID:           row.ID.String(),
			Name:         row.Name,
			ServiceID:    row.ServiceID.String(),
			FilterString: row.Filter,
			Actions:      actions,
		}
	}

	return rules, nil
}

// Create inserts the given rule, validating that the IntegrationKeyIDs and ServiceID
// are valid UUIDs. It also updates the service_rule_integration_keys pivot table
// with the rule's integration keys.
func (s *Store) Create(ctx context.Context, tx *sql.Tx, rule Rule) (*Rule, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	serviceID, err1 := validate.ParseUUID("ServiceID", rule.ServiceID)
	intKeyIDs, err2 := validate.ParseManyUUID("IntegrationKeyID", rule.IntegrationKeys, -1)
	if err = validate.Many(err1, err2); err != nil {
		return nil, err
	}

	actionsJson := pqtype.NullRawMessage{Valid: len(rule.Actions) > 0}
	if actionsJson.Valid {
		actionsJson.RawMessage, err = json.Marshal(rule.Actions)
		if err != nil {
			return nil, errors.Wrap(err, "marshal rule actions")
		}
	}

	ruleID, err := gadb.New(s.db).SvcRuleInsert(ctx, gadb.SvcRuleInsertParams{
		Name:      rule.Name,
		ServiceID: serviceID,
		Filter:    rule.FilterString,
		SendAlert: rule.SendAlert,
		Actions:   actionsJson,
	})
	if err != nil {
		return nil, err
	}

	err = gadb.New(tx).SvcRuleSetIntKeys(ctx, gadb.SvcRuleSetIntKeysParams{
		ServiceRuleID:     ruleID,
		IntegrationKeyIds: intKeyIDs,
	})
	if err != nil {
		return nil, err
	}

	rule.ID = ruleID.String()
	return &rule, nil
}

// Update updates the given rule, validating that the RuleID and IntegrationKeyIDs
// are valid UUIDs. It also updates the service_rule_integration_keys pivot table
// with the rule's integration keys.
func (s *Store) Update(ctx context.Context, tx *sql.Tx, rule Rule) (*Rule, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	ruleID, err1 := validate.ParseUUID("ServiceRuleID", rule.ID)
	intKeyIDs, err2 := validate.ParseManyUUID("IntegrationKeyID", rule.IntegrationKeys, -1)
	if err = validate.Many(err1, err2); err != nil {
		return nil, err
	}

	actionsJson := pqtype.NullRawMessage{Valid: len(rule.Actions) > 0}
	if actionsJson.Valid {
		actionsJson.RawMessage, err = json.Marshal(rule.Actions)
		if err != nil {
			return nil, errors.Wrap(err, "marshal rule actions")
		}
	}

	err = gadb.New(s.db).SvcRuleUpdate(ctx, gadb.SvcRuleUpdateParams{
		ID:        ruleID,
		Name:      rule.Name,
		Filter:    rule.FilterString,
		SendAlert: rule.SendAlert,
		Actions:   actionsJson,
	})
	if err != nil {
		return nil, err
	}

	err = gadb.New(tx).SvcRuleSetIntKeys(ctx, gadb.SvcRuleSetIntKeysParams{
		ServiceRuleID:     ruleID,
		IntegrationKeyIds: intKeyIDs,
	})
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

// Delete deletes the rule with the given ID.
func (s *Store) Delete(ctx context.Context, id string) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	ruleID, err := validate.ParseUUID("RuleID", id)
	if err != nil {
		return err
	}

	return gadb.New(s.db).SvcRuleDelete(ctx, ruleID)
}
