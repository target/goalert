package integrationkeyrule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation/validate"
)

type Store struct{}

type GoAlertTemplate struct {
	Summary string `json:"summary"`
	Details string `json:"details"`
	Dedup   string `json:"dedup"`
}

func (s *Store) FindManyByIntKey(ctx context.Context, dbtx gadb.DBTX, id string) ([]Rule, error) {
	ruleID, err := validate.ParseUUID("IntegrationKeyRuleID", id)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(dbtx).IntKeyRuleFindManyByIntKey(ctx, ruleID)
	if err != nil {
		return nil, errors.Wrap(err, "get rules for service")
	}

	rules := []Rule{}

	for _, row := range rows {
		var template GoAlertTemplate
		err := processTemplate(row.Msgtemplate, &template)
		if err != nil {
			defaultTemplate(&template)
		}

		rule := Rule{
			ID:             row.ID.String(),
			Name:           row.Name,
			IntegrationKey: row.IntegrationKeyID.String(),
			Filter:         row.Filter,
			Summary:        template.Summary,
			Details:        template.Details,
			Dedup:          template.Dedup,
			Action:         AlertActionType(row.Action),
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func processTemplate(templateMsg json.RawMessage, templateType interface{}) error {
	return json.Unmarshal(templateMsg, templateType)
}

func defaultTemplate(templateType interface{}) error {
	switch t := templateType.(type) {
	case *GoAlertTemplate:
		t.Summary = "Default Summary"
		t.Details = "Default Details"
		t.Dedup = "Default Dedup"
	default:
		return fmt.Errorf("unsupported template type: %T", templateType)
	}

	return nil
}
