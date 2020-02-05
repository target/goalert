package graphqlapp

import (
	context "context"
	"database/sql"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rule"

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

		for ruleIndex, inputRule := range input.Rules {
			r := rule.NewAlwaysActive(schedID, input.Target)
			if inputRule.Start != nil {
				r.Start = *inputRule.Start
			}
			if inputRule.End != nil {
				r.End = *inputRule.End
			}
			for i, v := range inputRule.WeekdayFilter {
				r.WeekdayFilter.SetDay(time.Weekday(i), v)
			}
			if ruleIndex < len(rules) {
				r.ID = rules[ruleIndex].ID
				err = errors.Wrap(m.RuleStore.UpdateTx(ctx, tx, r), "update rule")
			} else {
				_, err = m.RuleStore.CreateRuleTx(ctx, tx, r)
				err = errors.Wrap(err, "create rule")
			}
			if err != nil {
				return err
			}
		}

		if len(rules) > len(input.Rules) {
			toDelete := make([]string, len(rules)-len(input.Rules))
			for i, r := range rules[len(input.Rules):] {
				toDelete[i] = r.ID
			}
			err := errors.Wrap(m.RuleStore.DeleteManyTx(ctx, tx, toDelete), "delete old rules")
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err == nil, err
}
