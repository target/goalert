package graphqlapp

import (
	context "context"
	"database/sql"
	"strconv"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/validation"

	"github.com/pkg/errors"
)

type ScheduleRule App

func (a *App) ScheduleRule() graphql2.ScheduleRuleResolver { return (*ScheduleRule)(a) }
func (r *ScheduleRule) Target(ctx context.Context, raw *rule.Rule) (*assignment.RawTarget, error) {
	tgt := assignment.NewRawTarget(raw.Target)
	return &tgt, nil
}
func (r *ScheduleRule) WeekdayFilter(ctx context.Context, raw *rule.Rule) ([]bool, error) {
	var f [7]bool
	for i, v := range raw.WeekdayFilter {
		f[i] = v == 1
	}
	return f[:], nil
}

func (m *Mutation) UpdateScheduleTarget(ctx context.Context, input graphql2.ScheduleTargetInput) (bool, error) {
	var schedID string
	if input.ScheduleID != nil {
		schedID = *input.ScheduleID
	}
	if input.Target.Type == assignment.TargetTypeUser && input.Target.ID == "__current_user" {
		input.Target.ID = permission.UserID(ctx)
	}
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		_, err := m.ScheduleStore.FindOneForUpdate(ctx, tx, schedID) // lock schedule
		if err != nil {
			return errors.Wrap(err, "lock schedule")
		}

		rules, err := m.RuleStore.FindByTargetTx(ctx, tx, schedID, input.Target)
		if err != nil {
			return errors.Wrap(err, "fetch existing rules")
		}
		rulesByID := make(map[string]*rule.Rule, len(rules))
		for i := range rules {
			rulesByID[rules[i].ID] = &rules[i]
		}

		updated := make(map[string]bool, len(rules))
		for i, inputRule := range input.Rules {
			r := rule.NewAlwaysActive(schedID, input.Target)
			if inputRule.ID != nil {
				// doing an update
				if rulesByID[*inputRule.ID] == nil {
					return validation.NewFieldError("rules["+strconv.Itoa(i)+"]", "does not exist")
				}
				r = rulesByID[*inputRule.ID]
			}
			if inputRule.Start != nil {
				r.Start = *inputRule.Start
			}
			if inputRule.End != nil {
				r.End = *inputRule.End
			}
			for i, v := range inputRule.WeekdayFilter {
				r.WeekdayFilter.SetDay(time.Weekday(i), v)
			}

			if inputRule.ID != nil {
				updated[*inputRule.ID] = true
				err = errors.Wrap(m.RuleStore.UpdateTx(ctx, tx, r), "update rule")
			} else {
				_, err = m.RuleStore.CreateRuleTx(ctx, tx, r)
				err = errors.Wrap(err, "create rule")
			}
			if err != nil {
				return err
			}
		}

		toDelete := make([]string, 0, len(rules)-len(updated))
		for _, rule := range rules {
			if updated[rule.ID] {
				continue
			}
			toDelete = append(toDelete, rule.ID)
		}

		return errors.Wrap(m.RuleStore.DeleteManyTx(ctx, tx, toDelete), "delete old rules")
	})
	return err == nil, err
}
