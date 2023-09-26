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
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) *Store {
	return &Store{db: db}
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

	serviceUUID, err := validate.ParseUUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(s.db).SvcRuleFindManyByService(ctx, serviceUUID)
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

	serviceUUID, err1 := validate.ParseUUID("ServiceID", serviceID)
	keyUUID, err2 := validate.ParseUUID("IntegrationKeyID", integrationKeyID)
	if err = validate.Many(err1, err2); err != nil {
		return nil, err
	}

	rows, err := gadb.New(s.db).SvcRuleFindManyByIntKey(ctx, gadb.SvcRuleFindManyByIntKeyParams{
		ServiceID:        serviceUUID,
		IntegrationKeyID: keyUUID,
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

// InsertRule inserts the given rule, validating that the IntegrationKeyID and ServiceID
// are valid UUIDs.
func (s *Store) InsertRule(ctx context.Context, rule Rule) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	serviceUUID, err := validate.ParseUUID("ServiceID", rule.ServiceID)
	if err != nil {
		return err
	}

	actionsJson := pqtype.NullRawMessage{Valid: len(rule.Actions) > 0}
	if actionsJson.Valid {
		actionsJson.RawMessage, err = json.Marshal(rule.Actions)
		if err != nil {
			return errors.Wrap(err, "marshal rule actions")
		}
	}
	err = gadb.New(s.db).SvcRuleInsert(ctx, gadb.SvcRuleInsertParams{
		Name:      rule.Name,
		ServiceID: serviceUUID,
		Filter:    rule.FilterString,
		SendAlert: rule.SendAlert,
	})
	return errors.Wrap(err, "insert service rule")
}
